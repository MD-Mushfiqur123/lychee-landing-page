export class Lychee {
  constructor(config = {}) {
    this.host = (config.host || "http://localhost:11434").replace(/\/$/, "");
  }

  async _request(method, path, body = null) {
    const url = `${this.host}${path}`;
    const options = {
      method,
      headers: {
        "Content-Type": "application/json",
      },
    };
    if (body) {
      options.body = JSON.stringify(body);
    }
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error(`Lychee API error: ${response.status} ${response.statusText}`);
    }
    return response;
  }

  async generate(options) {
    const response = await this._request("POST", "/api/generate", options);
    if (options.stream) {
      return this._streamIterator(response);
    }
    return response.json();
  }

  async chat(options) {
    const response = await this._request("POST", "/api/chat", options);
    if (options.stream) {
      return this._streamIterator(response);
    }
    return response.json();
  }

  async list() {
    const response = await this._request("GET", "/api/tags");
    return response.json();
  }

  async show(model) {
    const response = await this._request("POST", "/api/show", { name: model });
    return response.json();
  }

  async create(options) {
    const response = await this._request("POST", "/api/create", options);
    if (options.stream) {
      return this._streamIterator(response);
    }
    return response.json();
  }

  async delete(model) {
    const response = await this._request("DELETE", "/api/delete", { name: model });
    return response.json();
  }

  async pull(options) {
    const response = await this._request("POST", "/api/pull", options);
    if (options.stream) {
      return this._streamIterator(response);
    }
    return response.json();
  }

  async push(options) {
    const response = await this._request("POST", "/api/push", options);
    if (options.stream) {
      return this._streamIterator(response);
    }
    return response.json();
  }

  async embed(options) {
    const response = await this._request("POST", "/api/embed", options);
    return response.json();
  }

  async ps() {
    const response = await this._request("GET", "/api/ps");
    return response.json();
  }

  async* _streamIterator(response) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop();
        for (const line of lines) {
          if (line.trim()) {
            yield JSON.parse(line);
          }
        }
      }
      if (buffer.trim()) {
        yield JSON.parse(buffer);
      }
    } finally {
      reader.releaseLock();
    }
  }
}

const defaultClient = new Lychee();

export default {
  Lychee,
  generate: defaultClient.generate.bind(defaultClient),
  chat: defaultClient.chat.bind(defaultClient),
  list: defaultClient.list.bind(defaultClient),
  show: defaultClient.show.bind(defaultClient),
  create: defaultClient.create.bind(defaultClient),
  delete: defaultClient.delete.bind(defaultClient),
  pull: defaultClient.pull.bind(defaultClient),
  push: defaultClient.push.bind(defaultClient),
  embed: defaultClient.embed.bind(defaultClient),
  ps: defaultClient.ps.bind(defaultClient),
};
