/** @jsx h */
import { h, Component, RenderableProps } from 'preact';
import b from 'bem-react-helper';

import { PROVIDER_NAMES, IS_STORAGE_AVAILABLE, IS_THIRD_PARTY } from '@app/common/constants';
import { requestDeletion } from '@app/utils/email';
import { getHandleClickProps } from '@app/common/accessibility';
import { User, AuthProvider, Sorting, Theme, PostInfo } from '@app/common/types';

import Dropdown, { DropdownItem } from '@app/components/dropdown';
import { Button } from '@app/components/button';
import { UserID } from './__user-id';
import { AnonymousLoginForm } from './__anonymous-login-form';
import { EmailLoginForm, EmailLoginFormConnected } from './__email-login-form';
import { StoreState } from '@app/store';

export interface Props {
  user: User | null;
  hiddenUsers: StoreState['hiddenUsers'];
  providers: (AuthProvider['name'])[];
  sort: Sorting;
  isCommentsDisabled: boolean;
  theme: Theme;
  postInfo: PostInfo;

  onSortChange(s: Sorting): Promise<void>;
  onSignIn(p: AuthProvider): Promise<User | null>;
  onSignOut(): Promise<void>;
  onCommentsEnable(): Promise<boolean>;
  onCommentsDisable(): Promise<boolean>;
  onBlockedUsersShow(): void;
  onBlockedUsersHide(): void;
}

interface State {
  isBlockedVisible: boolean;
  anonymousUsernameInputValue: string;
}

export class AuthPanel extends Component<Props, State> {
  emailLoginRef?: EmailLoginForm;

  constructor(props: Props) {
    super(props);

    this.state = {
      isBlockedVisible: false,
      anonymousUsernameInputValue: 'anon',
    };

    this.toggleBlockedVisibility = this.toggleBlockedVisibility.bind(this);
    this.toggleCommentsAvailability = this.toggleCommentsAvailability.bind(this);
    this.onSortChange = this.onSortChange.bind(this);
    this.onSignIn = this.onSignIn.bind(this);
    this.onEmailSignIn = this.onEmailSignIn.bind(this);
    this.handleAnonymousLoginFormSubmut = this.handleAnonymousLoginFormSubmut.bind(this);
    this.handleOAuthLogin = this.handleOAuthLogin.bind(this);
    this.toggleUserInfoVisibility = this.toggleUserInfoVisibility.bind(this);
    this.onEmailTitleClick = this.onEmailTitleClick.bind(this);
  }

  onEmailTitleClick() {
    this.emailLoginRef && this.emailLoginRef.focus();
  }

  onSortChange(e: Event) {
    if (this.props.onSortChange) {
      this.props.onSortChange((e.target! as HTMLOptionElement).value as Sorting);
    }
  }

  toggleBlockedVisibility() {
    if (!this.state.isBlockedVisible) {
      if (this.props.onBlockedUsersShow) this.props.onBlockedUsersShow();
    } else if (this.props.onBlockedUsersHide) this.props.onBlockedUsersHide();

    this.setState({ isBlockedVisible: !this.state.isBlockedVisible });
  }

  toggleCommentsAvailability() {
    if (this.props.isCommentsDisabled) {
      this.props.onCommentsEnable && this.props.onCommentsEnable();
    } else {
      this.props.onCommentsDisable && this.props.onCommentsDisable();
    }
  }

  toggleUserInfoVisibility() {
    const user = this.props.user;
    if (window.parent && user) {
      const data = JSON.stringify({ isUserInfoShown: true, user });
      window.parent.postMessage(data, '*');
    }
  }

  getUserTitle() {
    const { user } = this.props;
    return <span className="auth-panel__username">{user!.name}</span>;
  }

  /** wrapper function to handle both oauth and anonymous providers*/
  onSignIn(provider: AuthProvider) {
    this.props.onSignIn(provider);
  }

  onEmailSignIn(token: string) {
    return this.props.onSignIn({ name: 'email', token });
  }

  async handleAnonymousLoginFormSubmut(username: string) {
    this.onSignIn({ name: 'anonymous', username });
  }

  async handleOAuthLogin(e: MouseEvent | KeyboardEvent) {
    const p = (e.target as HTMLButtonElement).dataset.provider! as AuthProvider['name'];
    // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
    this.onSignIn({ name: p } as AuthProvider);
  }

  renderAuthorized = () => {
    const { user, onSignOut } = this.props;
    if (!user) return null;

    const isUserAnonymous = user && user.id.substr(0, 10) === 'anonymous_';

    return (
      <div className="auth-panel__column">
        You signed in as{' '}
        <Dropdown title={user.name} theme={this.props.theme}>
          <DropdownItem separator={!isUserAnonymous}>
            <UserID id={user.id} theme={this.props.theme} {...getHandleClickProps(this.toggleUserInfoVisibility)} />
          </DropdownItem>

          {!isUserAnonymous && (
            <DropdownItem>
              <Button kind="link" theme={this.props.theme} onClick={() => requestDeletion().then(onSignOut)}>
                Request my data removal
              </Button>
            </DropdownItem>
          )}
        </Dropdown>{' '}
        <Button className="auth-panel__sign-out" kind="link" theme={this.props.theme} onClick={onSignOut}>
          Sign out?
        </Button>
      </div>
    );
  };

  renderProvider = (provider: AuthProvider['name']) => {
    if (provider === 'anonymous') {
      return (
        <Dropdown title={PROVIDER_NAMES['anonymous']} titleClass="auth-panel__pseudo-link" theme={this.props.theme}>
          <DropdownItem>
            <AnonymousLoginForm
              onSubmit={this.handleAnonymousLoginFormSubmut}
              theme={this.props.theme}
              className="auth-panel__anonymous-login-form"
            />
          </DropdownItem>
        </Dropdown>
      );
    }
    if (provider === 'email') {
      return (
        <Dropdown
          title={PROVIDER_NAMES['email']}
          titleClass="auth-panel__pseudo-link"
          theme={this.props.theme}
          onTitleClick={this.onEmailTitleClick}
        >
          <DropdownItem>
            <EmailLoginFormConnected
              ref={ref => (this.emailLoginRef = ref ? ref.getWrappedInstance() : null)}
              onSignIn={this.onEmailSignIn}
              theme={this.props.theme}
              className="auth-panel__email-login-form"
            />
          </DropdownItem>
        </Dropdown>
      );
    }

    return (
      <span
        className="auth-panel__pseudo-link"
        data-provider={provider}
        // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
        {...getHandleClickProps(this.handleOAuthLogin)}
        role="link"
      >
        {PROVIDER_NAMES[provider]}
      </span>
    );
  };

  renderUnauthorized = () => {
    const { user, providers = [], postInfo } = this.props;
    if (user || !IS_STORAGE_AVAILABLE) return null;

    const signInMessage = postInfo.read_only ? 'Sign in using ' : 'Sign in to comment using ';

    return (
      <div className="auth-panel__column">
        {signInMessage}
        {providers.map((provider, i) => {
          const comma = i === 0 ? '' : i === providers.length - 1 ? ' or ' : ', ';

          return (
            <span>
              {comma} {this.renderProvider(provider)}
            </span>
          );
        })}
      </div>
    );
  };

  renderThirdPartyWarning = () => {
    if (IS_STORAGE_AVAILABLE || !IS_THIRD_PARTY) return null;
    return (
      <div className="auth-panel__column">
        Disable third-party cookies blocking to sign in or open comments in{' '}
        <a
          class="auth-panel__pseudo-link"
          href={`${window.location.origin}/web/comments.html${window.location.search}`}
          target="_blank"
        >
          new page
        </a>
      </div>
    );
  };

  renderCookiesWarning = () => {
    if (IS_STORAGE_AVAILABLE || IS_THIRD_PARTY) return null;
    return <div className="auth-panel__column">Allow cookies to sign in and comment</div>;
  };

  renderSettingsLabel = () => {
    return (
      <span
        className="auth-panel__pseudo-link auth-panel__admin-action"
        {...getHandleClickProps(() => this.toggleBlockedVisibility())}
        role="link"
      >
        {this.state.isBlockedVisible ? 'Hide' : 'Show'} settings
      </span>
    );
  };

  renderReadOnlySwitch = () => {
    const { isCommentsDisabled } = this.props;
    return (
      <span
        className="auth-panel__pseudo-link auth-panel__admin-action"
        {...getHandleClickProps(() => this.toggleCommentsAvailability())}
        role="link"
      >
        {isCommentsDisabled ? 'Enable' : 'Disable'} comments
      </span>
    );
  };

  renderSort = () => {
    const { sort } = this.props;
    const sortArray = getSortArray(sort);
    return (
      <span className="auth-panel__sort">
        Sort by{' '}
        <span className="auth-panel__select-label">
          {sortArray.find(x => 'selected' in x && x.selected!)!.label}
          <select className="auth-panel__select" onChange={this.onSortChange} onBlur={this.onSortChange}>
            {sortArray.map(sort => (
              <option value={sort.value} selected={sort.selected}>
                {sort.label}
              </option>
            ))}
          </select>
        </span>
      </span>
    );
  };

  render(props: RenderableProps<Props>, { isBlockedVisible }: State) {
    const {
      user,
      postInfo: { read_only },
      theme,
    } = props;
    const isAdmin = user && user.admin;
    const isSettingsLabelVisible = Object.keys(this.props.hiddenUsers).length > 0 || isAdmin || isBlockedVisible;

    return (
      <div className={b('auth-panel', {}, { theme, loggedIn: !!user })}>
        {this.renderAuthorized()}
        {this.renderUnauthorized()}
        {this.renderThirdPartyWarning()}
        {this.renderCookiesWarning()}
        <div className="auth-panel__column">
          {isSettingsLabelVisible && this.renderSettingsLabel()}

          {isSettingsLabelVisible && ' • '}

          {isAdmin && this.renderReadOnlySwitch()}

          {isAdmin && ' • '}

          {!isAdmin && read_only && <span className="auth-panel__readonly-label">Read-only</span>}

          {this.renderSort()}
        </div>
      </div>
    );
  }
}

function getSortArray(currentSort: Sorting) {
  const sortArray: {
    value: Sorting;
    label: string;
    selected?: boolean;
  }[] = [
    {
      value: '-score',
      label: 'Best',
    },
    {
      value: '+score',
      label: 'Worst',
    },
    {
      value: '-time',
      label: 'Newest',
    },
    {
      value: '+time',
      label: 'Oldest',
    },
    {
      value: '-active',
      label: 'Recently updated',
    },
    {
      value: '+active',
      label: 'Least recently updated',
    },
    {
      value: '-controversy',
      label: 'Most controversial',
    },
    {
      value: '+controversy',
      label: 'Least controversial',
    },
  ];

  return sortArray.map(sort => {
    if (sort.value === currentSort) {
      sort.selected = true;
    }

    return sort;
  });
}
