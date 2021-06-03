import { h, JSX, Component, FunctionComponent } from 'preact';
import b from 'bem-react-helper';
import { useSelector } from 'react-redux';
import { useIntl, defineMessages, FormattedMessage, IntlShape } from 'react-intl';

import { StoreState } from 'store';
import { Comment } from 'common/types';
import { fetchInfo } from 'store/user-info/actions';
import { parseQuery } from 'utils/parseQuery';
import { bindActions } from 'utils/actionBinder';
import { postMessageToParent } from 'utils/postMessage';
import { useActions } from 'hooks/useAction';
import { Avatar } from 'components/avatar';

import { LastCommentsList } from './last-comments-list';

const boundActions = bindActions({ fetchInfo });

const messages = defineMessages({
  unexpectedError: {
    id: 'user-info.unexpected-error',
    defaultMessage: 'Something went wrong',
  },
});

const user = parseQuery();

type Props = {
  comments: Comment[] | null;
} & typeof boundActions & { intl: IntlShape };

interface State {
  isLoading: boolean;
  error: string | null;
}

class UserInfo extends Component<Props, State> {
  state = { isLoading: true, error: null };

  componentWillMount(): void {
    if (!this.props.comments && this.state.isLoading) {
      this.props
        .fetchInfo(user.id)
        .then(() => {
          this.setState({ isLoading: false });
        })
        .catch(() => {
          this.setState({ isLoading: false, error: this.props.intl.formatMessage(messages.unexpectedError) });
        });
    }

    document.addEventListener('keydown', UserInfo.onKeyDown);
  }

  componentWillUnmount(): void {
    document.removeEventListener('keydown', UserInfo.onKeyDown);
  }

  render(): JSX.Element | null {
    const { comments = [] } = this.props;
    const { isLoading } = this.state;

    // TODO: handle
    if (!user) {
      return null;
    }

    return (
      <div className={b('root user-info', {})}>
        <Avatar className="user-info__avatar" url={user.picture} />
        <p className="user-info__title">
          <FormattedMessage
            id="user-info.last-comments"
            defaultMessage="Last comments by {userName}"
            values={{ userName: user.name }}
          />
        </p>
        <p className="user-info__id">{user.id}</p>
        {!!comments && <LastCommentsList isLoading={isLoading} comments={comments} />}
      </div>
    );
  }

  /**
   * Global on `keydown` handler which is set on component mount.
   * Listens for user's `esc` key press
   */
  static onKeyDown(e: KeyboardEvent): void {
    // ESCAPE key pressed
    if (e.keyCode === 27) {
      postMessageToParent({ profile: null });
    }
  }
}

const commentsSelector = (state: StoreState) => state.userComments[user.id];

export const ConnectedUserInfo: FunctionComponent = () => {
  const comments = useSelector(commentsSelector);
  const actions = useActions(boundActions);
  const intl = useIntl();

  return <UserInfo comments={comments} {...actions} intl={intl} />;
};
