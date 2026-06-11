// Type declarations for webview API functions

interface ImageData {
  filename: string;
  path: string;
  dataURL: string; // base64 encoded file data
}

interface MenuItem {
  label: string;
  enabled?: boolean;
  separator?: boolean;
}

interface WebviewAPI {
  selectFile: () => Promise<ImageData | null>;
  selectMultipleFiles: () => Promise<ImageData[] | null>;
  selectModelsDirectory: () => Promise<string | null>;
  selectWorkingDirectory: () => Promise<string | null>;
}

declare global {
  interface Window {
    webview?: WebviewAPI;
    drag?: () => void;
    doubleClick?: () => void;
    menu: (items: MenuItem[]) => Promise<string | null>;
    LYCHEE_TOOLS?: boolean;
    LYCHEE_WEBSEARCH?: boolean;
  }

  namespace JSX {
    interface IntrinsicElements {
      input: React.DetailedHTMLProps<
        React.InputHTMLAttributes<HTMLInputElement> & {
          webkitdirectory?: string;
          directory?: string;
        },
        HTMLInputElement
      >;
    }
  }

  interface File {
    readonly webkitRelativePath: string;
  }
}

export type { ImageData, WebviewAPI, ContextMenuItem, ContextMenuResult };

declare module "lychee/browser" {
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
}
