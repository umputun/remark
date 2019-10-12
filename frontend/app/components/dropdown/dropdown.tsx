/** @jsx h */
import { Component, h, RenderableProps } from 'preact';
import b from 'bem-react-helper';

import { Button } from '@app/components/button';
import { Theme } from '@app/common/types';
import { sleep } from '@app/utils/sleep';
import { DropdownItem } from '@app/components/dropdown/index';

interface Props {
  title: string | JSX.Element;
  titleClass?: string;
  heading?: string;
  isActive?: boolean;
  onTitleClick?: () => void;
  mix?: string;
  theme: Theme;
  onOpen?: (root: HTMLDivElement) => {};
  onClose?: (root: HTMLDivElement) => {};
  selectableItems?: string[];
  activeSelectableItem?: number;
  onDropdownItemClick?: () => void;
}

interface State {
  isActive: boolean;
  contentTranslateX: number;
  selectableItems?: string[];
  activeSelectableItem: number;
  filter?: string;
}

export default class Dropdown extends Component<Props, State> {
  rootNode?: HTMLDivElement;

  constructor(props: Props) {
    super(props);

    const { isActive, selectableItems } = this.props;
    let { activeSelectableItem } = this.props;

    if (activeSelectableItem === undefined) {
      activeSelectableItem = 0;
    }

    this.state = {
      isActive: isActive || false,
      contentTranslateX: 0,
      activeSelectableItem,
      selectableItems,
    };

    this.onOutsideClick = this.onOutsideClick.bind(this);
    this.receiveMessage = this.receiveMessage.bind(this);
    this.__onOpen = this.__onOpen.bind(this);
    this.__onClose = this.__onClose.bind(this);
    this.open = this.open.bind(this);
    this.close = this.close.bind(this);
    this.onDropdownItemHover = this.onDropdownItemHover.bind(this);
  }

  selectNextSelectableItem() {
    const { selectableItems, activeSelectableItem } = this.state;

    if (!selectableItems) return;

    const itemsLength = selectableItems.length;
    const firstItem = 0;

    let newActiveSelectableItem = activeSelectableItem + 1;

    if (newActiveSelectableItem >= itemsLength) {
      newActiveSelectableItem = firstItem;
    }

    this.setState({
      activeSelectableItem: newActiveSelectableItem,
    });
  }

  selectPreviousSelectableItem() {
    const { selectableItems, activeSelectableItem } = this.state;

    if (!selectableItems) return;

    const itemsLength = selectableItems.length;
    const lastItem = itemsLength - 1;

    let newActiveSelectableItem = activeSelectableItem - 1;

    if (newActiveSelectableItem < 0) {
      newActiveSelectableItem = lastItem;
    }

    this.setState({
      activeSelectableItem: newActiveSelectableItem,
    });
  }

  setFilter(filter?: string) {
    this.setState({
      filter,
    });
  }

  getSelectedItem() {
    const { selectableItems, activeSelectableItem } = this.state;

    if (!selectableItems || activeSelectableItem === undefined) return;

    return selectableItems[activeSelectableItem];
  }

  filterSelectableList(): void {
    const { filter } = this.state;
    if (!filter) return;

    const list = this.props.selectableItems || [];

    const newSelectableList = list.filter(emoji => {
      if (filter) {
        if (emoji.indexOf(filter) === -1) return false;
      }
      return true;
    });

    this.setState({
      selectableItems: newSelectableList,
    });
  }

  generateList(list: string[]) {
    const items = list.map((emoji, index) => {
      return (
        <DropdownItem
          index={index}
          onFocus={this.onDropdownItemHover}
          onMouseOver={this.onDropdownItemHover}
          active={index === this.state.activeSelectableItem}
          onDropdownItemClick={this.props.onDropdownItemClick}
        >
          {emoji}
        </DropdownItem>
      );
    });

    if (items.length === 0) {
      return <DropdownItem>No such emoji</DropdownItem>;
    }

    return items;
  }

  onTitleClick() {
    const isActive = !this.state.isActive;
    const contentTranslateX = isActive ? this.state.contentTranslateX : 0;
    this.setState(
      {
        contentTranslateX,
        isActive,
      },
      async () => {
        await this.__adjustDropDownContent();
        if (isActive) {
          this.__onOpen();
          this.props.onOpen && this.props.onOpen(this.rootNode!);
        } else {
          this.__onClose();
          this.props.onClose && this.props.onClose(this.rootNode!);
        }

        if (this.props.onTitleClick) {
          this.props.onTitleClick();
        }
      }
    );
  }

  storedDocumentHeight: string | null = null;
  storedDocumentHeightSet: boolean = false;
  checkInterval: number | undefined = undefined;

  __onOpen() {
    const isChildOfDropDown = (() => {
      if (!this.rootNode) return false;
      let parent = this.rootNode.parentElement!;
      while (parent !== document.body) {
        if (parent.classList.contains('dropdown')) return true;
        parent = parent.parentElement!;
      }
      return false;
    })();
    if (isChildOfDropDown) return;

    this.storedDocumentHeight = document.body.style.minHeight;
    this.storedDocumentHeightSet = true;

    let prevDcBottom: number | null = null;

    this.checkInterval = window.setInterval(() => {
      if (!this.rootNode || !this.state.isActive) return;
      const windowHeight = window.innerHeight;
      const dcBottom = (() => {
        const dc = Array.from(this.rootNode.children).find(c => c.classList.contains('dropdown__content'));
        if (!dc) return 0;
        const rect = dc.getBoundingClientRect();
        return window.scrollY + Math.abs(rect.top) + dc.scrollHeight + 10;
      })();
      if (prevDcBottom === null && dcBottom <= windowHeight) return;
      if (dcBottom !== prevDcBottom) {
        prevDcBottom = dcBottom;
        document.body.style.minHeight = dcBottom + 'px';
      }
    }, 100);
  }

  /**
   * Force open dropdown if close
   */
  open() {
    if (this.state.isActive) return;

    this.setState({
      isActive: true,
    });
    this.__adjustDropDownContent().then(() => this.__onOpen());
  }

  __onClose() {
    const { selectableItems } = this.props;

    window.clearInterval(this.checkInterval);
    if (this.storedDocumentHeightSet) {
      document.body.style.minHeight = this.storedDocumentHeight;
    }

    this.setState({
      activeSelectableItem: 0,
      filter: undefined,
      selectableItems,
    });
  }

  /**
   * Force close dropdown if open
   */
  close() {
    if (!this.state.isActive) return;

    this.setState({
      isActive: false,
    });
    this.__adjustDropDownContent();
    this.__onClose();
  }

  async __adjustDropDownContent() {
    if (!this.rootNode) return;
    const dc = this.rootNode.querySelector<HTMLDivElement>('.dropdown__content');
    if (!dc) return;
    await sleep(10);
    const rect = dc.getBoundingClientRect();
    if (rect.left > 0) {
      const wWindow = window.innerWidth;
      if (rect.right <= wWindow) return;
      const delta = rect.right - wWindow;
      const max = Math.min(rect.left, delta);
      this.setState({
        contentTranslateX: -max,
      });
      return;
    }
    this.setState({
      contentTranslateX: -rect.left,
    });
  }

  receiveMessage(e: { data: string | object }) {
    try {
      const data = typeof e.data === 'string' ? JSON.parse(e.data) : e.data;

      if (!data.clickOutside) return;
      if (!this.state.isActive) return;
      this.setState(
        {
          contentTranslateX: 0,
          isActive: false,
        },
        () => {
          this.__onClose();
          this.props.onClose && this.props.onClose(this.rootNode!);
        }
      );
    } catch (e) {}
  }

  onOutsideClick(e: MouseEvent) {
    if (!this.rootNode || this.rootNode.contains(e.target as Node) || !this.state.isActive) return;
    this.setState(
      {
        contentTranslateX: 0,
        isActive: false,
      },
      () => {
        this.__onClose();
        this.props.onClose && this.props.onClose(this.rootNode!);
      }
    );
  }

  componentDidMount() {
    document.addEventListener('click', this.onOutsideClick);

    window.addEventListener('message', this.receiveMessage);
  }

  componentWillUnmount() {
    document.removeEventListener('click', this.onOutsideClick);

    window.removeEventListener('message', this.receiveMessage);
  }

  onDropdownItemHover(e: Event) {
    const target = e.target;

    if (target instanceof HTMLElement) {
      const { id } = target.dataset;

      if (id !== undefined) {
        this.setState({
          activeSelectableItem: +id,
        });
      }
    }
  }

  render(props: RenderableProps<Props>, { isActive }: State) {
    let { children } = props;
    const { title, titleClass, heading, mix } = props;

    {
      if (this.state.selectableItems) {
        children = this.generateList(this.state.selectableItems);
      }
    }

    return (
      <div className={b('dropdown', { mix }, { theme: props.theme, active: isActive })} ref={r => (this.rootNode = r)}>
        <Button
          aria-haspopup="listbox"
          aria-expanded={isActive && 'true'}
          mix="dropdown__title"
          type="button"
          onClick={() => this.onTitleClick()}
          theme="light"
          className={titleClass}
        >
          {title}
        </Button>

        <div
          className="dropdown__content"
          tabIndex={-1}
          role="listbox"
          style={{ transform: `translateX(${this.state.contentTranslateX}px)` }}
        >
          {heading && <div className="dropdown__heading">{heading}</div>}
          <div className="dropdown__items">{children}</div>
        </div>
      </div>
    );
  }
}
