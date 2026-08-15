from dataclasses import dataclass, field


@dataclass
class ChatRequest:
    query: str


@dataclass
class ChatResponse:
    answer: str
    context: list[str] = field(default_factory=list)
