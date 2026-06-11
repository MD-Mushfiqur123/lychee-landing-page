export interface ModelDetails {
  parent_model?: string;
  format?: string;
  family?: string;
  families?: string[];
  parameter_size?: string;
  quantization_level?: string;
}

export interface ModelResponse {
  name: string;
  modified_at?: string;
  size?: number;
  digest?: string;
  details?: ModelDetails;
}

export class Lychee {
  constructor(config?: { host?: string });
  host: string;
  generate(options: any): Promise<any>;
  chat(options: any): Promise<any>;
  list(): Promise<{ models: ModelResponse[] }>;
  show(options: { model: string }): Promise<any>;
  create(options: any): Promise<any>;
  delete(model: string): Promise<any>;
  pull(options: any): Promise<any>;
  push(options: any): Promise<any>;
  embed(options: any): Promise<any>;
  ps(): Promise<any>;
}

declare const defaultClient: {
  Lychee: typeof Lychee;
  generate: (options: any) => Promise<any>;
  chat: (options: any) => Promise<any>;
  list: () => Promise<{ models: ModelResponse[] }>;
  show: (options: { model: string }) => Promise<any>;
  create: (options: any) => Promise<any>;
  delete: (model: string) => Promise<any>;
  pull: (options: any) => Promise<any>;
  push: (options: any) => Promise<any>;
  embed: (options: any) => Promise<any>;
  ps: () => Promise<any>;
};

export default defaultClient;
