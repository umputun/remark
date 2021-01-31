global.Headers = class HeadersMock extends Headers implements Headers {
  private headers = new Map();

  constructor(headers?: HeadersInit) {
    super(headers);

    if (typeof headers === 'object' && headers !== null) {
      const entries = Object.entries(headers) as [string, string][];

      if (entries.length > 0) {
        entries.forEach(([k, v]) => {
          this.append(k, v);
        });
      }
    }
  }

  append(key: string, value: string) {
    this.headers.set(key, value);
  }
  set(key: string, value: string) {
    this.headers.set(key, value);
  }
  has(key: string) {
    return this.headers.has(key);
  }
  get(key: string) {
    return this.headers.get(key) || null;
  }
  delete(key: string) {
    this.headers.delete(key);
  }
  forEach(callbackfn: (value: string, key: string, parent: Headers) => void, thisArg?: unknown) {
    this.headers.forEach((value, key) => {
      callbackfn.call(thisArg || this, value, key, this);
    });
  }
};
