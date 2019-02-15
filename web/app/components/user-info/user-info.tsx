/** @jsx h */
import { h, Component, RenderableProps } from 'preact';
import b from 'bem-react-helper';
import { connect } from 'preact-redux';

import { StoreState, StoreDispatch } from '@app/store';
import { Comment } from '@app/common/types';
import { fetchInfo } from '@app/store/user-info/actions';
import { userInfo } from '@app/common/user-info-settings';

import LastCommentsList from './last-comments-list';
import { AvatarIcon } from '../avatar-icon';

interface Props {
  comments: Comment[] | null;
  fetchInfo: () => Promise<Comment[] | null>;
}

interface State {
  isLoading: boolean;
  error: string | null;
}

class UserInfo extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { isLoading: true, error: null };
  }

  async componentWillMount(): Promise<void> {
    if (!this.props.comments && this.state.isLoading) {
      this.props
        .fetchInfo()
        .then(() => {
          this.setState({ isLoading: false });
        })
        .catch(() => {
          this.setState({ isLoading: false, error: 'Something went wrong' });
        });
    }

    document.addEventListener('keydown', this.globalOnKeyDown);
  }

  componentWillUnmount(): void {
    document.removeEventListener('keydown', this.globalOnKeyDown);
  }

  globalOnKeyDown(e: KeyboardEvent): void {
    // ESCAPE key pressed
    if (e.keyCode === 27) {
      const data = JSON.stringify({ isUserInfoShown: false });
      window.parent.postMessage(data, '*');
    }
  }

  render(props: RenderableProps<Props>, state: State): JSX.Element | null {
    const user = userInfo;
    const { comments = [] } = props;

    // TODO: handle
    if (!user) {
      return null;
    }

    return (
      <div className={b('user-info', {})}>
        <AvatarIcon mix="user-info__avatar" picture={user.picture} />
        <p className="user-info__title">Last comments by {user.name}</p>
        <p className="user-info__id">{user.id}</p>

        {!!comments && <LastCommentsList isLoading={state.isLoading} comments={comments} />}
      </div>
    );
  }
}

const mapDispatchToProps = (dispatch: StoreDispatch) => ({
  fetchInfo: () => dispatch(fetchInfo()),
});

export const ConnectedUserInfo = connect(
  (
    state: StoreState
  ): {
    comments: Comment[] | null;
  } => ({
    comments: state.userComments![userInfo.id!],
  }),
  mapDispatchToProps
)(UserInfo);
