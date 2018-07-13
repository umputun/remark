/** @jsx h */
import { h, Component } from 'preact';
import api from 'common/api';

import {
  BASE_URL,
  NODE_ID,
  COMMENT_NODE_CLASSNAME_PREFIX,
  DEFAULT_SORT,
  LS_SORT_KEY,
  MAX_SHOWN_ROOT_COMMENTS,
} from 'common/constants';
import { siteId, url, maxShownComments } from 'common/settings';
import store from 'common/store';

import AuthPanel from 'components/auth-panel';
import BlockedUsers from 'components/blocked-users';
import Comment from 'components/comment';
import Input from 'components/input';
import Preloader from 'components/preloader';
import Thread from 'components/thread';

const IS_MOBILE = /Android|webOS|iPhone|iPad|iPod|Opera Mini|Windows Phone/i.test(navigator.userAgent);

export default class Root extends Component {
  constructor(props) {
    super(props);

    let sort;

    try {
      sort = localStorage.getItem(LS_SORT_KEY) || DEFAULT_SORT;
    } catch (e) {
      sort = DEFAULT_SORT;
    }

    this.state = {
      isLoaded: false,
      isCommentsListLoading: false,
      user: {},
      sort,
      commentsShown: maxShownComments || MAX_SHOWN_ROOT_COMMENTS,
    };

    this.addComment = this.addComment.bind(this);
    this.replaceComment = this.replaceComment.bind(this);
    this.onSignIn = this.onSignIn.bind(this);
    this.onSignOut = this.onSignOut.bind(this);
    this.onBlockedUsersShow = this.onBlockedUsersShow.bind(this);
    this.onBlockedUsersHide = this.onBlockedUsersHide.bind(this);
    this.onCommentsDisable = this.onCommentsDisable.bind(this);
    this.onCommentsEnable = this.onCommentsEnable.bind(this);
    this.onSortChange = this.onSortChange.bind(this);
    this.onUnblockSomeone = this.onUnblockSomeone.bind(this);
    this.checkUrlHash = this.checkUrlHash.bind(this);
    this.showMore = this.showMore.bind(this);
  }

  componentWillMount() {
    store.onUpdate('comments', comments => this.setState({ comments }));
    store.onUpdate('info', info => this.setState({ info }));
  }

  componentDidMount() {
    const { sort } = this.state;

    api.getConfig().then(config => {
      store.set('config', config);
      this.setState({ config });
    });

    Promise.all([
      api
        .getUser()
        .then(data => store.set('user', data))
        .catch(() => store.set('user', {})),
      api
        .getPostComments({ sort, url })
        .then(({ comments = [], info = {} } = {}) => {
          store.set('comments', comments);
          store.set('info', info);
        })
        .catch(() => store.set('comments', [])),
    ]).finally(() => {
      this.setState({
        isLoaded: true,
        user: store.get('user'),
      });

      setTimeout(this.checkUrlHash);
      window.addEventListener('hashchange', this.checkUrlHash);
    });
  }

  findComment(comments, commentId) {
    let comment;
    for (let i = 0; i < comments.length; i += 1) {
      comment = comments[i];
      if (comment.comment.id === commentId) {
        return comment;
      }
      if (comment.replies != null && comment.replies.length > 0) {
        comment = this.findComment(comment.replies, commentId);
        if (comment) {
          return comment;
        }
      }
    }
  }

  checkUrlHash(e) {
    const hash = e ? `#${e.newURL.split('#')[1]}` : window.location.hash;

    if (hash.indexOf(`#${COMMENT_NODE_CLASSNAME_PREFIX}`) === 0) {
      if (e) {
        e.preventDefault();
      }

      const commentEl = document.querySelector(hash);
      if (commentEl) {
        setTimeout(() => {
          window.parent.postMessage(JSON.stringify({ scrollTo: commentEl.getBoundingClientRect().top }), '*');
        }, 500);

        return;
      }

      if (IS_MOBILE) {
        const commentId = hash.replace(`#${COMMENT_NODE_CLASSNAME_PREFIX}`, '');
        let comment = this.findComment(this.state.comments, commentId);
        if (comment == null) {
          return;
        }

        while (comment.comment.pid.length !== 0) {
          comment = this.findComment(this.state.comments, comment.comment.pid);
        }

        const commentPos = this.state.comments.indexOf(comment) + 1;
        if (commentPos > this.state.commentsShown) {
          this.setState({
            commentsShown: commentPos,
          });

          setTimeout(() => {
            this.checkUrlHash(e);
          }, 100);

          return;
        }
      }
    }
  }

  onSignOut() {
    api.logOut().then(() => {
      store.set('user', {});
      this.setState({ user: {} });
    });
  }

  onSignIn(provider) {
    const newWindow = window.open(
      `${BASE_URL}/auth/${provider}/login?from=${encodeURIComponent(
        location.origin + location.pathname + '?selfClose'
      )}&site=${siteId}`
    );

    let secondsPass = 0;
    const checkMsDelay = 300;
    const checkInterval = setInterval(() => {
      let shouldProceed;
      secondsPass += checkMsDelay;
      try {
        shouldProceed = newWindow.closed || secondsPass > 30000;
      } catch (e) {}

      if (shouldProceed) {
        clearInterval(checkInterval);

        api
          .getUser()
          .then(user => {
            store.set('user', user);
            this.setState({ user });
          })
          .catch(() => {}); // TODO: we need to handle it and write error to user
      }
    }, checkMsDelay);
  }

  onBlockedUsersShow() {
    api.getBlocked().then(bannedUsers => {
      this.setState({ bannedUsers, isBlockedVisible: true });
    });
  }

  onCommentsEnable() {
    api.enableComments(siteId, url).then(() => {
      const info = store.get('info');
      info.read_only = false;
      this.setState({ info });
    });
  }

  onCommentsDisable() {
    api.disableComments(siteId, url).then(() => {
      const info = store.get('info');
      info.read_only = true;
      this.setState({ info });
    });
  }

  onBlockedUsersHide() {
    const { wasSomeoneUnblocked, sort } = this.state;

    // if someone was unblocked let's reload comments
    if (wasSomeoneUnblocked) {
      api.getPostComments({ sort, url }).then(({ comments, info } = {}) => {
        store.set('comments', comments);
        store.set('info', info);
      });
    }

    this.setState({
      isBlockedVisible: false,
      wasSomeoneUnblocked: false,
    });
  }

  onSortChange(sort) {
    if (sort === this.state.sort) return;

    this.setState({ sort, isCommentsListLoading: true });

    try {
      localStorage.setItem(LS_SORT_KEY, sort);
    } catch (e) {
      // can't save; ignore it
    }

    api
      .getPostComments({ sort, url })
      .then(({ comments, info } = {}) => {
        store.set('comments', comments);
        store.set('info', info);
      })
      .finally(() => {
        this.setState({ isCommentsListLoading: false });
      });
  }

  onUnblockSomeone() {
    this.setState({ wasSomeoneUnblocked: true });
  }

  addComment(comment) {
    store.addComment(comment);
  }

  replaceComment(comment) {
    store.replaceComment(comment);
  }

  showMore() {
    this.setState({
      commentsShown: this.state.commentsShown + MAX_SHOWN_ROOT_COMMENTS,
    });
  }

  render(
    props,
    {
      config = {},
      comments = [],
      info = {},
      user,
      sort,
      isLoaded,
      isBlockedVisible,
      isCommentsListLoading,
      bannedUsers,
      commentsShown,
    }
  ) {
    if (!isLoaded) {
      return (
        <div id={NODE_ID}>
          <div className="root">
            <Preloader mix="root__preloader" />
          </div>
        </div>
      );
    }

    // TODO: i think we should do it on backend
    const pinnedComments = store.getPinnedComments();
    const isGuest = !Object.keys(user).length;
    const isCommentsDisabled = info != null && info.read_only === true;

    return (
      <div id={NODE_ID}>
        <div className="root">
          <AuthPanel
            user={user}
            sort={sort}
            providers={config.auth_providers}
            isCommentsDisabled={isCommentsDisabled}
            onSignIn={this.onSignIn}
            onSignOut={this.onSignOut}
            onBlockedUsersShow={this.onBlockedUsersShow}
            onBlockedUsersHide={this.onBlockedUsersHide}
            onCommentsEnable={this.onCommentsEnable}
            onCommentsDisable={this.onCommentsDisable}
            onSortChange={this.onSortChange}
          />

          {!isBlockedVisible && (
            <div className="root__main">
              {!isGuest &&
                !isCommentsDisabled && <Input mix="root__input" mods={{ type: 'main' }} onSubmit={this.addComment} />}

              {!!pinnedComments.length && (
                <div className="root__pinned-comments" role="region" aria-label="Pinned comments">
                  {pinnedComments.map(comment => (
                    <Comment data={comment} mods={{ level: 0, disabled: true }} mix="root__pinned-comment" />
                  ))}
                </div>
              )}

              {!!comments.length &&
                !isCommentsListLoading && (
                  <div className="root__threads" role="list">
                    {(IS_MOBILE ? comments.slice(0, commentsShown) : comments).map(thread => (
                      <Thread
                        key={thread.comment.id}
                        mix="root__thread"
                        mods={{ level: 0 }}
                        data={thread}
                        isCommentsDisabled={isCommentsDisabled}
                        onReply={this.addComment}
                        onEdit={this.replaceComment}
                      />
                    ))}

                    {commentsShown < comments.length &&
                      IS_MOBILE && (
                        <button className="root__show-more" onClick={this.showMore}>
                          Show more
                        </button>
                      )}
                  </div>
                )}

              {isCommentsListLoading && (
                <div className="root__threads" role="list">
                  <Preloader mix="root__preloader" />
                </div>
              )}
            </div>
          )}

          {isBlockedVisible && (
            <div className="root__main">
              <BlockedUsers users={bannedUsers} onUnblock={this.onUnblockSomeone} />
            </div>
          )}

          <p className="root__copyright" role="contentinfo">
            Powered by{' '}
            <a href="https://remark42.com/" className="root__copyright-link">
              Remark42
            </a>
          </p>
        </div>
      </div>
    );
  }
}
