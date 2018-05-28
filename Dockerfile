FROM umputun/baseimage:buildgo-latest as build-backend

ARG COVERALLS_TOKEN
ARG CI
ARG TRAVIS
ARG TRAVIS_BRANCH
ARG TRAVIS_COMMIT
ARG TRAVIS_JOB_ID
ARG TRAVIS_JOB_NUMBER
ARG TRAVIS_OS_NAME
ARG TRAVIS_PULL_REQUEST
ARG TRAVIS_PULL_REQUEST_SHA
ARG TRAVIS_REPO_SLUG
ARG TRAVIS_TAG

WORKDIR /go/src/github.com/umputun/remark

ADD app /go/src/github.com/umputun/remark/app
ADD vendor /go/src/github.com/umputun/remark/vendor
ADD .git /go/src/github.com/umputun/remark/.git

RUN cd app && go test ./...

RUN gometalinter --disable-all --deadline=300s --vendor --enable=vet --enable=vetshadow --enable=golint \
    --enable=staticcheck --enable=ineffassign --enable=goconst --enable=errcheck --enable=unconvert \
    --enable=deadcode  --enable=gosimple --enable=gas --exclude=test --exclude=mock --exclude=vendor ./...

RUN mkdir -p target && /script/coverage.sh

RUN if [ -z "$COVERALLS_TOKEN" ] ; then \
    echo coverall not enabled ; \
    else goveralls -coverprofile=.cover/cover.out -service=travis-ci -repotoken $COVERALLS_TOKEN; fi

RUN go build -o remark -ldflags "-X main.revision=$(git rev-parse --abbrev-ref HEAD)-$(git describe --abbrev=7 --always --tags)-$(date +%Y%m%d-%H:%M:%S) -s -w" ./app


FROM node:9.4-alpine as build-frontend

ADD web /srv/web
RUN apk add --no-cache --update git
RUN \
    cd /srv/web && \
    npm i && npm run build && \
    rm -rf ./node_modules


FROM umputun/baseimage:app-latest

WORKDIR /srv

ADD scripts/import-disqus.sh /srv/import-disqus.sh
ADD scripts/restore-backup.sh /srv/restore-backup.sh
ADD scripts/migrate-data.sh /srv/migrate-data.sh
ADD scripts/create-backup.sh /srv/create-backup.sh

ADD start.sh /srv/start.sh

RUN chmod +x /srv/start.sh /srv/import-disqus.sh /srv/restore-backup.sh /srv/migrate-data.sh /srv/create-backup.sh

COPY --from=build-backend /go/src/github.com/umputun/remark/remark /srv/
COPY --from=build-frontend /srv/web/public/ /srv/web
RUN chown -R app:app /srv

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1

CMD ["/srv/start.sh"]
ENTRYPOINT ["/init.sh"]
