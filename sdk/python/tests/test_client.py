import unittest
from unittest.mock import patch, MagicMock
import json

from lychee import Client

class TestLycheeClient(unittest.TestCase):
    def setUp(self):
        self.client = Client()

    @patch('requests.request')
    def test_generate(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"model": "llama3", "response": "hello", "done": True}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.generate("llama3", "hi")
        self.assertEqual(res["response"], "hello")
        mock_request.assert_called_once_with(
            "POST",
            "http://localhost:11434/api/generate",
            json={"model": "llama3", "prompt": "hi", "stream": False, "raw": False}
        )

    @patch('requests.request')
    def test_generate_stream(self, mock_request):
        mock_response = MagicMock()
        mock_response.iter_lines.return_value = [
            b'{"response": "hel", "done": false}',
            b'{"response": "lo", "done": true}'
        ]
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        generator = self.client.generate("llama3", "hi", stream=True)
        chunks = list(generator)
        self.assertEqual(len(chunks), 2)
        self.assertEqual(chunks[0]["response"], "hel")
        self.assertEqual(chunks[1]["response"], "lo")

    @patch('requests.request')
    def test_chat(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"model": "llama3", "message": {"role": "assistant", "content": "hi"}}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        messages = [{"role": "user", "content": "hello"}]
        res = self.client.chat("llama3", messages)
        self.assertEqual(res["message"]["content"], "hi")

    @patch('requests.request')
    def test_chat_stream(self, mock_request):
        mock_response = MagicMock()
        mock_response.iter_lines.return_value = [
            b'{"message": {"role": "assistant", "content": "h"}, "done": false}',
            b'{"message": {"role": "assistant", "content": "i"}, "done": true}'
        ]
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        messages = [{"role": "user", "content": "hello"}]
        generator = self.client.chat("llama3", messages, stream=True)
        chunks = list(generator)
        self.assertEqual(len(chunks), 2)
        self.assertEqual(chunks[0]["message"]["content"], "h")
        self.assertEqual(chunks[1]["message"]["content"], "i")

    @patch('requests.request')
    def test_list(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"models": [{"name": "llama3:latest"}]}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.list()
        self.assertEqual(len(res["models"]), 1)
        self.assertEqual(res["models"][0]["name"], "llama3:latest")

    @patch('requests.request')
    def test_show(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"modelfile": "FROM llama3"}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.show("llama3")
        self.assertEqual(res["modelfile"], "FROM llama3")

    @patch('requests.request')
    def test_create(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"status": "success"}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.create("my-model", modelfile="FROM llama3")
        self.assertEqual(res["status"], "success")

    @patch('requests.request')
    def test_delete(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"status": "deleted"}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.delete("my-model")
        self.assertEqual(res["status"], "deleted")

    @patch('requests.request')
    def test_pull(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"status": "pulling"}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.pull("llama3")
        self.assertEqual(res["status"], "pulling")

    @patch('requests.request')
    def test_push(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"status": "pushing"}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.push("llama3")
        self.assertEqual(res["status"], "pushing")

    @patch('requests.request')
    def test_embed(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"embeddings": [[0.1, 0.2]]}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.embed("llama3", "hello")
        self.assertEqual(len(res["embeddings"]), 1)
        self.assertEqual(res["embeddings"][0], [0.1, 0.2])

    @patch('requests.request')
    def test_ps(self, mock_request):
        mock_response = MagicMock()
        mock_response.json.return_value = {"models": []}
        mock_response.status_code = 200
        mock_request.return_value = mock_response

        res = self.client.ps()
        self.assertEqual(res["models"], [])

if __name__ == '__main__':
    unittest.main()
