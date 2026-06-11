import json
import requests
from typing import Any, Dict, Generator, List, Union, Optional

class Client:
    def __init__(self, host: Optional[str] = None):
        self.host = (host or "http://localhost:11434").rstrip("/")

    def _request(self, method: str, path: str, **kwargs) -> requests.Response:
        url = f"{self.host}{path}"
        return requests.request(method, url, **kwargs)

    def _stream(self, method: str, path: str, **kwargs) -> Generator[Dict[str, Any], None, None]:
        url = f"{self.host}{path}"
        response = requests.request(method, url, stream=True, **kwargs)
        response.raise_for_status()
        for line in response.iter_lines():
            if line:
                yield json.loads(line.decode("utf-8"))

    def generate(
        self,
        model: str,
        prompt: str,
        system: Optional[str] = None,
        template: Optional[str] = None,
        options: Optional[Dict[str, Any]] = None,
        stream: bool = False,
        raw: bool = False,
    ) -> Union[Dict[str, Any], Generator[Dict[str, Any], None, None]]:
        payload = {
            "model": model,
            "prompt": prompt,
            "stream": stream,
            "raw": raw,
        }
        if system:
            payload["system"] = system
        if template:
            payload["template"] = template
        if options:
            payload["options"] = options

        if stream:
            return self._stream("POST", "/api/generate", json=payload)
        else:
            response = self._request("POST", "/api/generate", json=payload)
            response.raise_for_status()
            return response.json()

    def chat(
        self,
        model: str,
        messages: List[Dict[str, Any]],
        options: Optional[Dict[str, Any]] = None,
        stream: bool = False,
    ) -> Union[Dict[str, Any], Generator[Dict[str, Any], None, None]]:
        payload = {
            "model": model,
            "messages": messages,
            "stream": stream,
        }
        if options:
            payload["options"] = options

        if stream:
            return self._stream("POST", "/api/chat", json=payload)
        else:
            response = self._request("POST", "/api/chat", json=payload)
            response.raise_for_status()
            return response.json()

    def list(self) -> Dict[str, Any]:
        response = self._request("GET", "/api/tags")
        response.raise_for_status()
        return response.json()

    def show(self, model: str) -> Dict[str, Any]:
        payload = {"name": model}
        response = self._request("POST", "/api/show", json=payload)
        response.raise_for_status()
        return response.json()

    def create(
        self,
        model: str,
        modelfile: Optional[str] = None,
        path: Optional[str] = None,
        stream: bool = False,
    ) -> Union[Dict[str, Any], Generator[Dict[str, Any], None, None]]:
        payload = {"name": model}
        if modelfile:
            payload["modelfile"] = modelfile
        if path:
            payload["path"] = path

        if stream:
            return self._stream("POST", "/api/create", json=payload)
        else:
            response = self._request("POST", "/api/create", json=payload)
            response.raise_for_status()
            return response.json()

    def delete(self, model: str) -> Dict[str, Any]:
        payload = {"name": model}
        response = self._request("DELETE", "/api/delete", json=payload)
        response.raise_for_status()
        return response.json()

    def pull(
        self,
        model: str,
        stream: bool = False,
    ) -> Union[Dict[str, Any], Generator[Dict[str, Any], None, None]]:
        payload = {"name": model}
        if stream:
            return self._stream("POST", "/api/pull", json=payload)
        else:
            response = self._request("POST", "/api/pull", json=payload)
            response.raise_for_status()
            return response.json()

    def push(
        self,
        model: str,
        stream: bool = False,
    ) -> Union[Dict[str, Any], Generator[Dict[str, Any], None, None]]:
        payload = {"name": model}
        if stream:
            return self._stream("POST", "/api/push", json=payload)
        else:
            response = self._request("POST", "/api/push", json=payload)
            response.raise_for_status()
            return response.json()

    def embed(
        self,
        model: str,
        input: Union[str, List[str]],
        options: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        payload = {
            "model": model,
            "input": input,
        }
        if options:
            payload["options"] = options
        response = self._request("POST", "/api/embed", json=payload)
        response.raise_for_status()
        return response.json()

    def ps(self) -> Dict[str, Any]:
        response = self._request("GET", "/api/ps")
        response.raise_for_status()
        return response.json()
