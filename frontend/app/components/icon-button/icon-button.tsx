import clsx from 'clsx';
import { h, RenderableProps, JSX } from 'preact';

import styles from './icon-button.module.css';

type Props = JSX.HTMLAttributes<HTMLButtonElement> &
  RenderableProps<{
    className?: string;
  }>;

export function IconButton({ title, children, className }: Props) {
  return (
    <button title={title} className={clsx('icon-button', className, styles.root)}>
      {children}
    </button>
  );
}
