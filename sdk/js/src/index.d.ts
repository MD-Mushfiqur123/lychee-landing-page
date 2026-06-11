export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface ComposeCondition {
  contains?: string;
  not_contains?: string;
  min_length?: number;
  max_length?: number;
  always?: boolean;
}

export interface ComposeStep {
  model: string;
  prompt: string;
  options?: Record<string, any>;
  timeout_sec?: number;
  fallback_model?: string;
  parallel?: ComposeStep[];
  condition?: ComposeCondition;
  skip_on_error?: boolean;
}

export interface ComposeRequest {
  input: string;
  steps: ComposeStep[];
  stream?: boolean;
}

export class LycheeError extends Error {
  statusCode?: number;
  constructor(message: string, statusCode?: number);
}

export class Lychee {
  constructor(baseUrl?: string);
  baseUrl: string;

  chat(
    model: string,
    message: string,
    options?: {
      system?: string;
      history?: Message[];
      stream?: boolean;
      inferenceOptions?: Record<string, any>;
    }
  ): Promise<any> | AsyncGenerator<any, void, unknown>;

  generate(
    model: string,
    prompt: string,
    options?: {
      stream?: boolean;
      inferenceOptions?: Record<string, any>;
    }
  ): Promise<any> | AsyncGenerator<any, void, unknown>;

  compose(params: ComposeRequest): Promise<any> | AsyncGenerator<any, void, unknown>;

  messages(
    model: string,
    messages: Message[],
    options?: {
      maxTokens?: number;
      system?: string;
      stream?: boolean;
    }
  ): Promise<any> | AsyncGenerator<any, void, unknown>;

  chatCompletions(
    model: string,
    messages: any[],
    options?: {
      stream?: boolean;
      [key: string]: any;
    }
  ): Promise<any> | AsyncGenerator<any, void, unknown>;

  listModels(): Promise<any[]>;

  pull(
    model: string,
    options?: {
      stream?: boolean;
    }
  ): Promise<any> | AsyncGenerator<any, void, unknown>;

  show(model: string): Promise<any>;

  structured(
    model: string,
    prompt: string,
    schema: any,
    options?: {
      maxRetries?: number;
      options?: Record<string, any>;
    }
  ): Promise<any>;

  listConversations(): Promise<any[]>;
  getConversation(id: string): Promise<any>;
  deleteConversation(id: string): Promise<any>;

  createRoute(
    name: string,
    endpoints: Array<{ host: string; model?: string }>,
    strategy?: 'round_robin' | 'random' | 'least_loaded'
  ): Promise<any>;
  listRoutes(): Promise<any[]>;
  deleteRoute(name: string): Promise<any>;

  isRunning(): Promise<boolean>;
}
