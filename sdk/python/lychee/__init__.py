from lychee.client import Client

_default_client = Client()

generate = _default_client.generate
chat = _default_client.chat
list = _default_client.list
show = _default_client.show
create = _default_client.create
delete = _default_client.delete
pull = _default_client.pull
push = _default_client.push
embed = _default_client.embed
ps = _default_client.ps

__all__ = [
    "Client",
    "generate",
    "chat",
    "list",
    "show",
    "create",
    "delete",
    "pull",
    "push",
    "embed",
    "ps",
]
