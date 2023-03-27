export class APIError extends Error {
  constructor(
    private _res: Response,
    private _status: number,
    private _message?: string
  ) {
    super(_message ?? "unknown");
  }

  get response() {
    return this._res;
  }

  get status() {
    return this._res.status;
  }

  get code() {
    return this._status;
  }
}
