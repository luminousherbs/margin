export interface MarginSession {
  authenticated: boolean;
  did?: string;
  handle?: string;
  accessJwt?: string;
  refreshJwt?: string;
}

export interface TextSelector {
  type?: string;
  exact: string;
  prefix?: string;
  suffix?: string;
}

export interface Annotation {
  uri?: string;
  id?: string;
  cid?: string;
  type?: 'Annotation' | 'Bookmark' | 'Highlight';
  body?: { value: string };
  text?: string;
  target?: {
    source?: string;
    selector?: TextSelector;
  };
  selector?: TextSelector;
  color?: string;
  tags?: string[];
  created?: string;
  createdAt?: string;
  creator?: Author;
  author?: Author;
}

export interface Author {
  did?: string;
  handle?: string;
  displayName?: string;
  avatar?: string;
}

export interface Bookmark {
  uri?: string;
  id?: string;
  source?: string;
  url?: string;
  title?: string;
  description?: string;
  image?: string;
  tags?: string[];
  createdAt?: string;
  target?: {
    source?: string;
    title?: string;
  };
}

export interface Highlight {
  uri?: string;
  id?: string;
  target?: {
    source?: string;
    selector?: TextSelector;
  };
  color?: string;
  tags?: string[];
  title?: string;
  createdAt?: string;
}

export interface Collection {
  uri?: string;
  id?: string;
  name: string;
  description?: string;
  icon?: string;
  createdAt?: string;
  itemCount?: number;
}

export const DEFAULT_API_URL = 'https://margin.at';
export const APP_URL = 'https://margin.at';
