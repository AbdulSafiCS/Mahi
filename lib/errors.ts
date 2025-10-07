export class ApiError extends Error {
  status: number;
  code?: string;
  details?: any;

  constructor(opts: {
    status: number;
    message: string;
    code?: string;
    details?: any;
  }) {
    super(opts.message);
    this.name = "ApiError";
    this.status = opts.status;
    this.code = opts.code;
    this.details = opts.details;
  }
}
