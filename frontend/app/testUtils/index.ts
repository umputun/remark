import 'jest-extended';
import { StaticStore } from '@app/common/static_store';

require('document-register-element/pony')(window);

beforeEach(() => {
  StaticStore.config = {
    admin_email: 'admin@remark42.com',
    admins: ['admin'],
    auth_providers: ['dev', 'google'],
    critical_score: -15,
    low_score: -5,
    edit_duration: 300,
    max_comment_size: 3000,
    max_image_size: 5000,
    positive_score: false,
    readonly_age: 100,
    version: 'jest-test',
  };
});

export function createDomContainer(setup: (domContainer: HTMLElement) => void): void {
  let domContainer: HTMLElement | null = null;
  beforeAll(() => {
    domContainer = document.createElement('div');
    (document.body || document.documentElement).appendChild(domContainer);
    setup(domContainer);
  });

  beforeEach(() => {
    domContainer!.innerHTML = '';
  });

  afterAll(() => {
    domContainer!.parentNode!.removeChild(domContainer!);
    domContainer = null;
  });
}
