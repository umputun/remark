package api

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/token"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"
	"github.com/go-pkgz/rest/cache"
	"github.com/hashicorp/go-multierror"

	"github.com/umputun/remark/backend/app/notify"
	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/image"
	"github.com/umputun/remark/backend/app/store/service"
)

type private struct {
	dataService      privStore
	cache            cache.LoadingCache
	readOnlyAge      int
	commentFormatter *store.CommentFormatter
	imageService     *image.Service
	emailService     notify.VerificationDestination
	notifyService    *notify.Service
	authenticator    *auth.Service
	remarkURL        string
}

type privStore interface {
	Create(comment store.Comment) (commentID string, err error)
	EditComment(locator store.Locator, commentID string, req service.EditRequest) (comment store.Comment, err error)
	Vote(req service.VoteReq) (comment store.Comment, err error)
	Get(locator store.Locator, commentID string, user store.User) (store.Comment, error)
	User(siteID, userID string, limit, skip int, user store.User) ([]store.Comment, error)
	GetUserDetail(locator store.Locator, userID string, detail string) (string, error)
	SetUserDetail(locator store.Locator, userID string, detail string, value string, delete bool) (string, error)
	ValidateComment(c *store.Comment) error
	IsVerified(siteID string, userID string) bool
	IsReadOnly(locator store.Locator) bool
	IsBlocked(siteID string, userID string) bool
	Info(locator store.Locator, readonlyAge int) (store.PostInfo, error)
}

// POST /comment - adds comment, resets all immutable fields
func (s *private) createCommentCtrl(w http.ResponseWriter, r *http.Request) {

	comment := store.Comment{}
	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &comment); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't bind comment", rest.ErrDecode)
		return
	}

	user := rest.MustGetUserInfo(r)
	if user.ID != "admin" && user.SiteID != comment.Locator.SiteID {
		rest.SendErrorJSON(w, r, http.StatusForbidden,
			errors.New("site mismatch, not allowed to post to "+comment.Locator.SiteID), "invalid site",
			rest.ErrCommentValidation)
		return
	}

	comment.PrepareUntrusted() // clean all fields user not supposed to set
	comment.User = user
	comment.User.IP = strings.Split(r.RemoteAddr, ":")[0]

	comment.Orig = comment.Text // original comment text, prior to md render
	if err := s.dataService.ValidateComment(&comment); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "invalid comment", rest.ErrCommentValidation)
		return
	}
	comment = s.commentFormatter.Format(comment)

	// check if user blocked
	if s.dataService.IsBlocked(comment.Locator.SiteID, comment.User.ID) {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"), "user blocked", rest.ErrUserBlocked)
		return
	}

	if s.isReadOnly(comment.Locator) {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"), "old post, read-only", rest.ErrReadOnly)
		return
	}

	id, err := s.dataService.Create(comment)
	if err == service.ErrRestrictedWordsFound {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "invalid comment", rest.ErrCommentValidation)
		return
	}
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't save comment", rest.ErrInternal)
		return
	}

	// dataService modifies comment
	finalComment, err := s.dataService.Get(comment.Locator, id, rest.GetUserOrEmpty(r))
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't load created comment", rest.ErrInternal)
		return
	}
	s.cache.Flush(cache.Flusher(comment.Locator.SiteID).
		Scopes(comment.Locator.URL, lastCommentsScope, comment.User.ID, comment.Locator.SiteID))

	if s.notifyService != nil {
		s.notifyService.Submit(finalComment)
	}

	log.Printf("[DEBUG] created commend %+v", finalComment)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &finalComment)
}

// PUT /comment/{id}?site=siteID&url=post-url - update comment
func (s *private) updateCommentCtrl(w http.ResponseWriter, r *http.Request) {

	edit := struct {
		Text    string
		Summary string
		Delete  bool
	}{}

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &edit); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't bind comment", rest.ErrDecode)
		return
	}

	user := rest.MustGetUserInfo(r)
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}
	id := chi.URLParam(r, "id")

	log.Printf("[DEBUG] update comment %s", id)

	var currComment store.Comment
	var err error
	if currComment, err = s.dataService.Get(locator, id, rest.GetUserOrEmpty(r)); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't find comment", rest.ErrCommentNotFound)
		return
	}

	if currComment.User.ID != user.ID {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"),
			"can not edit comments for other users", rest.ErrNoAccess)
		return
	}

	editReq := service.EditRequest{
		Text:    s.commentFormatter.FormatText(edit.Text),
		Orig:    edit.Text,
		Summary: edit.Summary,
		Delete:  edit.Delete,
	}

	res, err := s.dataService.EditComment(locator, id, editReq)
	if err == service.ErrRestrictedWordsFound {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "invalid comment", rest.ErrCommentValidation)
		return
	}

	if err != nil {
		code := parseError(err, rest.ErrCommentRejected)
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't update comment", code)
		return
	}

	s.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.SiteID, locator.URL, lastCommentsScope, user.ID))
	render.JSON(w, r, res)
}

// GET /user?site=siteID - returns user info
func (s *private) userInfoCtrl(w http.ResponseWriter, r *http.Request) {
	user := rest.MustGetUserInfo(r)
	if siteID := r.URL.Query().Get("site"); siteID != "" {
		user.Verified = s.dataService.IsVerified(siteID, user.ID)
	}

	render.JSON(w, r, user)
}

// PUT /vote/{id}?site=siteID&url=post-url&vote=1 - vote for/against comment
func (s *private) voteCtrl(w http.ResponseWriter, r *http.Request) {
	user := rest.MustGetUserInfo(r)
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}
	id := chi.URLParam(r, "id")
	log.Printf("[DEBUG] vote for comment %s", id)

	vote := r.URL.Query().Get("vote") == "1"

	if s.isReadOnly(locator) {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"), "old post, read-only", rest.ErrReadOnly)
		return
	}

	// check if user blocked
	if s.dataService.IsBlocked(locator.SiteID, user.ID) {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"), "user blocked", rest.ErrUserBlocked)
		return
	}

	req := service.VoteReq{
		Locator:   locator,
		CommentID: id,
		UserID:    user.ID,
		UserIP:    strings.Split(r.RemoteAddr, ":")[0],
		Val:       vote,
	}
	comment, err := s.dataService.Vote(req)
	if err != nil {
		code := parseError(err, rest.ErrVoteRejected)
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't vote for comment", code)
		return
	}
	s.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.URL, comment.User.ID))
	render.JSON(w, r, R.JSON{"id": comment.ID, "score": comment.Score})
}

// sendConfirmation gets address and siteID from query, makes confirmation token and sends it to user.
// GET /email?site=site&address=someone@example.com
func (s *private) sendConfirmation(w http.ResponseWriter, r *http.Request) {
	if s.emailService == nil {
		rest.SendErrorJSON(w, r, http.StatusNotImplemented, errors.New("unavailable"), "email notifications disabled", rest.ErrNotificationsDisabled)
		return
	}
	user := rest.MustGetUserInfo(r)
	address := r.URL.Query().Get("address")
	siteID := r.URL.Query().Get("site")
	if address == "" {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, errors.New("missing parameter"), "address parameter is required", rest.ErrInternal)
		return
	}
	existingAddress, _ := s.dataService.GetUserDetail(store.Locator{SiteID: siteID}, user.ID, "email")
	if address == existingAddress {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, errors.New("wrong request"), "address is already verified", rest.ErrInternal)
		return
	}
	claims := token.Claims{
		Handshake: &token.Handshake{ID: user.ID + "::" + address},
		StandardClaims: jwt.StandardClaims{
			Audience:  r.URL.Query().Get("site"),
			ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
			NotBefore: time.Now().Add(-1 * time.Minute).Unix(),
			Issuer:    "remark42",
		},
	}

	tkn, err := s.authenticator.TokenService().Token(claims)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusForbidden, err, "failed to make verification token", rest.ErrInternal)
		return
	}

	if err := s.emailService.SendVerification(r.Context(), notify.VerificationRequest{Locator: store.Locator{SiteID: siteID}, User: user.Name, Email: address, Token: tkn}); err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "failed to send confirmation", rest.ErrInternal)
		return
	}

	render.JSON(w, r, R.JSON{"user": user, "address": address})
}

// emailCtrl uses provided token parameter (generated by sendConfirmation) to set email and add it to user token
// PUT /email?site=siteID&tkn=token
func (s *private) emailCtrl(w http.ResponseWriter, r *http.Request) {
	tkn := r.URL.Query().Get("tkn")
	if tkn == "" {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, errors.New("missing parameter"), "token parameter is required", rest.ErrInternal)
		return
	}
	user := rest.MustGetUserInfo(r)
	locator := store.Locator{SiteID: r.URL.Query().Get("site")}

	confClaims, err := s.authenticator.TokenService().Parse(tkn)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusForbidden, err, "failed to verify confirmation token", rest.ErrInternal)
		return
	}

	if s.authenticator.TokenService().IsExpired(confClaims) {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("expired"), "failed to verify confirmation token", rest.ErrInternal)
		return
	}

	elems := strings.Split(confClaims.Handshake.ID, "::")
	if len(elems) != 2 || elems[0] != user.ID {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, errors.New(confClaims.Handshake.ID), "invalid handshake token", rest.ErrInternal)
		return
	}
	address := elems[1]

	log.Printf("[DEBUG] set email for user %s", user.ID)

	val, err := s.dataService.SetUserDetail(locator, user.ID, "email", address, false)
	if err != nil {
		code := parseError(err, rest.ErrInternal)
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't set email for user", code)
		return
	}

	// update User.Email from the token
	claims, _, err := s.authenticator.TokenService().Get(r)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusForbidden, err, "failed to verify confirmation token", rest.ErrInternal)
		return
	}
	claims.User.Email = address
	if _, err = s.authenticator.TokenService().Set(w, claims); err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "failed to set token", rest.ErrInternal)
		return
	}
	render.JSON(w, r, R.JSON{"updated": true, "address": val})
}

// DELETE /email?site=siteID - removes user's email
func (s *private) deleteEmailCtrl(w http.ResponseWriter, r *http.Request) {
	user := rest.MustGetUserInfo(r)
	locator := store.Locator{SiteID: r.URL.Query().Get("site")}
	log.Printf("[DEBUG] remove email for user %s", user.ID)

	_, err := s.dataService.SetUserDetail(locator, user.ID, "email", "", true)
	if err != nil {
		code := parseError(err, rest.ErrInternal)
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't delete email for user", code)
		return
	}
	// clean User.Email from the token
	claims, _, err := s.authenticator.TokenService().Get(r)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusForbidden, err, "failed to verify confirmation token", rest.ErrInternal)
		return
	}
	if claims.User.Email != "" {
		claims.User.Email = ""
		if _, err = s.authenticator.TokenService().Set(w, claims); err != nil {
			rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "failed to set token", rest.ErrInternal)
			return
		}
	}
	render.JSON(w, r, R.JSON{"deleted": true})
}

// GET /userdata?site=siteID - exports all data about the user as a json with user info and list of all comments
func (s *private) userAllDataCtrl(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site")
	user := rest.MustGetUserInfo(r)
	userB, err := json.Marshal(&user)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't marshal user info", rest.ErrInternal)
		return
	}

	exportFile := fmt.Sprintf("%s-%s-%s.json.gz", siteID, user.ID, time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment;filename="+exportFile)
	gzWriter := gzip.NewWriter(w)
	defer func() {
		if e := gzWriter.Close(); e != nil {
			log.Printf("[WARN] can't close gzip writer, %s", e)
		}
	}()

	write := func(val []byte) error {
		_, e := gzWriter.Write(val)
		return e
	}

	var merr error
	merr = multierror.Append(merr, write([]byte(`{"info": `)))     // send user prefix
	merr = multierror.Append(merr, write(userB))                   // send user info
	merr = multierror.Append(merr, write([]byte(`, "comments":`))) // send comments prefix

	// get comments in 100 in each paginated request
	for i := 0; i < 100; i++ {
		comments, errUser := s.dataService.User(siteID, user.ID, 100, i*100, rest.GetUserOrEmpty(r))
		if errUser != nil {
			rest.SendErrorJSON(w, r, http.StatusInternalServerError, errUser, "can't get user comments", rest.ErrInternal)
			return
		}
		b, errUser := json.Marshal(comments)
		if errUser != nil {
			rest.SendErrorJSON(w, r, http.StatusInternalServerError, errUser, "can't marshal user comments", rest.ErrInternal)
			return
		}

		merr = multierror.Append(merr, write(b))
		if len(comments) != 100 {
			break
		}
	}

	merr = multierror.Append(merr, write([]byte(`}`)))
	if merr.(*multierror.Error).ErrorOrNil() != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, merr, "can't write user info", rest.ErrInternal)
		return
	}

}

// POST /deleteme?site_id=site - requesting delete of all user info
// makes jwt with user info and sends it back as a part of json response
func (s *private) deleteMeCtrl(w http.ResponseWriter, r *http.Request) {
	user := rest.MustGetUserInfo(r)
	siteID := r.URL.Query().Get("site")

	claims := token.Claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  siteID,
			Issuer:    "remark42",
			ExpiresAt: time.Now().AddDate(0, 3, 0).Unix(),
			NotBefore: time.Now().Add(-1 * time.Minute).Unix(),
		},
		User: &token.User{
			ID:   user.ID,
			Name: user.Name,
			Attributes: map[string]interface{}{
				"delete_me": true, // prevents this token from being used for login
			},
		},
	}

	tokenStr, err := s.authenticator.TokenService().Token(claims)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't make token", rest.ErrInternal)
		return
	}

	link := fmt.Sprintf("%s/web/deleteme.html?token=%s", s.remarkURL, tokenStr)
	render.JSON(w, r, R.JSON{"site": siteID, "user_id": user.ID, "token": tokenStr, "link": link})
}

// POST /image - save image with form request
func (s *private) savePictureCtrl(w http.ResponseWriter, r *http.Request) {
	user := rest.MustGetUserInfo(r)

	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil { // 5M max memory, if bigger will make a file
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't parse multipart form", rest.ErrDecode)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't get image file from the request", rest.ErrInternal)
		return
	}
	defer func() { _ = file.Close() }()

	id, err := s.imageService.Save(header.Filename, user.ID, file)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't save image", rest.ErrInternal)
		return
	}

	render.JSON(w, r, R.JSON{"id": id})
}

func (s *private) isReadOnly(locator store.Locator) bool {
	if s.readOnlyAge > 0 {
		// check RO by age
		if info, e := s.dataService.Info(locator, s.readOnlyAge); e == nil && info.ReadOnly {
			return true
		}
	}
	return s.dataService.IsReadOnly(locator) // ro manually
}
