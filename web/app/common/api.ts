import { siteId, url } from './settings';
import { BASE_URL } from './constants';
import { Config, Comment, Tree, User, BlockedUser, Sorting, Provider, BlockTTL } from './types';
import fetcher from './fetcher';

/* common */

export const logIn = (provider: Provider) => {
  return new Promise<User | null>((resolve, reject) => {
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
        shouldProceed = (newWindow && newWindow.closed) || secondsPass > 30000;
      } catch (e) {}

      if (shouldProceed) {
        clearInterval(checkInterval);

        getUser()
          .then(user => {
            resolve(user);
          })
          .catch(() => {
            reject(new Error('User logIn Error'));
          });
      }
    }, checkMsDelay);
  });
};

export const logOut = (): Promise<void> => fetcher.get({ url: `/auth/logout`, overriddenApiBase: '' });

export const getConfig = (): Promise<Config> => fetcher.get(`/config`);

// TODO: looks like we can get url from settings here and below
export const getPostComments = (sort: Sorting): Promise<Tree> =>
  fetcher.get(`/find?site=${siteId}&url=${url}&sort=${sort}&format=tree`);

export const getLastComments = ({ siteId, max }: { siteId: string; max: number }): Promise<Comment[]> =>
  fetcher.get(`/last/${max}?site=${siteId}`);

export const getCommentsCount = (siteId: string, urls: string[]): Promise<{ url: string; count: number }[]> =>
  fetcher.post({
    url: `/counts?site=${siteId}`,
    body: urls,
  });

export const getComment = ({ id }: { id: Comment['id'] }): Promise<Comment> => fetcher.get(`/id/${id}?url=${url}`);

export const getUserComments = ({
  userId,
  limit,
}: {
  userId: User['id'];
  limit: number;
}): Promise<{
  comments: Comment[];
  count: number;
}> => fetcher.get(`/comments?user=${userId}&limit=${limit}`);

export const putCommentVote = ({ id, value }: { id: Comment['id']; value: number }): Promise<void> =>
  fetcher.put({
    url: `/vote/${id}?url=${url}&vote=${value}`,
    withCredentials: true,
  });

export const addComment = ({
  title,
  text,
  pid,
}: {
  title: string;
  text: string;
  pid?: Comment['id'];
}): Promise<Comment> =>
  fetcher.post({
    url: '/comment',
    body: {
      title,
      text,
      locator: {
        site: siteId,
        url,
      },
      ...(pid ? { pid } : {}),
    },
    withCredentials: true,
  });

export const updateComment = ({ text, id }: { text: string; id: Comment['id'] }): Promise<Comment> =>
  fetcher.put({
    url: `/comment/${id}?url=${url}`,
    body: {
      text,
    },
    withCredentials: true,
  });

export const getPreview = (text: string): Promise<string> =>
  fetcher.post({
    url: '/preview',
    body: {
      text,
    },
    withCredentials: true,
  });

export const getUser = (): Promise<User | null> =>
  fetcher
    .get<User | null>({
      url: '/user',
      withCredentials: true,
    })
    .catch(() => null);

/* GDPR */

export const deleteMe = (): Promise<{
  user_id: string;
  link: string;
}> =>
  fetcher.post({
    url: `/deleteme?site=${siteId}`,
  });

export const approveDeleteMe = (token: string): Promise<void> =>
  fetcher.get({
    url: `/admin/deleteme?token=${token}`,
  });

/* admin */
export const pinComment = (id: Comment['id']): Promise<void> =>
  fetcher.put({
    url: `/admin/pin/${id}?url=${url}&pin=1`,
    withCredentials: true,
  });

export const unpinComment = (id: Comment['id']): Promise<void> =>
  fetcher.put({
    url: `/admin/pin/${id}?url=${url}&pin=0`,
    withCredentials: true,
  });

export const setVerifyStatus = ({ id }: { id: User['id'] }): Promise<void> =>
  fetcher.put({
    url: `/admin/verify/${id}?verified=1`,
    withCredentials: true,
  });

export const removeVerifyStatus = ({ id }: { id: User['id'] }): Promise<void> =>
  fetcher.put({
    url: `/admin/verify/${id}?verified=0`,
    withCredentials: true,
  });

export const removeComment = ({ id }: { id: Comment['id'] }) =>
  fetcher.delete({
    url: `/admin/comment/${id}?url=${url}`,
    withCredentials: true,
  });

export const removeMyComment = ({ id }: { id: Comment['id'] }): Promise<void> =>
  fetcher.put({
    url: `/comment/${id}?url=${url}`,
    body: {
      delete: true,
    },
    withCredentials: true,
  });

export const blockUser = ({
  id,
  ttl,
}: {
  id: User['id'];
  ttl: BlockTTL;
}): Promise<{
  block: boolean;
  site_id: string;
  user_id: string;
}> =>
  fetcher.put({
    url: ttl === 'permanently' ? `/admin/user/${id}?block=1` : `/admin/user/${id}?block=1&ttl=${ttl}`,
    withCredentials: true,
  });

export const unblockUser = ({
  id,
}: {
  id: User['id'];
}): Promise<{
  block: boolean;
  site_id: string;
  user_id: string;
}> =>
  fetcher.put({
    url: `/admin/user/${id}?block=0`,
    withCredentials: true,
  });

export const getBlocked = (): Promise<BlockedUser[]> =>
  fetcher.get({
    url: '/admin/blocked',
    withCredentials: true,
  });

export const disableComments = (): Promise<void> =>
  fetcher.put({
    url: `/admin/readonly?site=${siteId}&url=${url}&ro=1`,
    withCredentials: true,
  });

export const enableComments = (): Promise<void> =>
  fetcher.put({
    url: `/admin/readonly?site=${siteId}&url=${url}&ro=0`,
    withCredentials: true,
  });

export default {
  logIn,
  logOut,
  getConfig,
  getPostComments,
  getLastComments,
  getCommentsCount,
  getComment,
  getUserComments,
  putCommentVote,
  addComment,
  updateComment,
  removeMyComment,
  getUser,
  getPreview,

  pinComment,
  unpinComment,
  setVerifyStatus,
  removeVerifyStatus,
  removeComment,
  blockUser,
  unblockUser,
  getBlocked,
  disableComments,
  enableComments,
};
