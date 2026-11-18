from dataclasses import dataclass
from typing import Protocol, Literal


@dataclass(frozen=True)
class Message:
    role: Literal["assistant", "system", "user"]
    content: str


class LLMProvider(Protocol):
    def generate(
        self,
        messages: list[Messages],
    ) -> str: ...
