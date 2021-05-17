import { render } from '@testing-library/preact';
import { h } from 'preact';
import createMockStore from 'redux-mock-store';
import { Middleware } from 'redux';
import { Provider } from 'react-redux';
import { IntlProvider } from 'react-intl';

import type { User } from 'common/types';
import enMessages from 'locales/en.json';

import { AuthPanel, Props } from './auth-panel';

const DefaultProps = {
  postInfo: {
    read_only: false,
    url: 'https://example.com',
    count: 3,
  },
  hiddenUsers: {},
} as Props;

const initialStore = {
  user: null,
  theme: 'light',
  comments: {
    sort: '-score',
  },
  provider: { name: 'google' },
} as const;

const mockStore = createMockStore([] as Middleware[]);

describe('<AuthPanel />', () => {
  const createContainer = (
    props: Props = DefaultProps,
    store: ReturnType<typeof mockStore> = mockStore(initialStore)
  ) =>
    render(
      <IntlProvider locale="en" messages={enMessages}>
        <Provider store={store}>
          <AuthPanel {...props} />
        </Provider>
      </IntlProvider>
    );

  describe('For not authorized : null', () => {
    it('should not render settings if there is no hidden users', () => {
      const { container } = createContainer({
        ...DefaultProps,
        user: null,
        postInfo: { ...DefaultProps.postInfo, read_only: true },
      } as Props);

      expect(container.querySelector('.auth-panel__admin-action')).not.toBeInTheDocument();
    });

    it('should render settings if there is some hidden users', () => {
      const { container } = createContainer({
        ...DefaultProps,
        user: null,
        postInfo: { ...DefaultProps.postInfo, read_only: true },
        hiddenUsers: { hidden_joe: {} as User },
      } as Props);

      expect(container.querySelector('.auth-panel__admin-action')).toHaveTextContent('Show settings');
    });
  });

  describe('For authorized user', () => {
    it('should render info about current user', () => {
      const { container } = createContainer({
        ...DefaultProps,
        user: { id: 'john', name: 'John' },
      } as Props);

      expect(container.querySelectorAll('.auth-panel__column')).toHaveLength(2);
      expect(container.querySelector('.auth-panel__column')?.textContent).toContain('You logged in as John');
    });
  });
  describe('For admin user', () => {
    it('should render admin action', () => {
      const { container } = createContainer({
        ...DefaultProps,
        user: { id: 'test', admin: true, name: 'John' },
      } as Props);

      expect(container.querySelector('.auth-panel__admin-action')).toHaveTextContent('Show settings');
    });
  });
});
