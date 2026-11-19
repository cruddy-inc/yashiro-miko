from dataclasses import dataclass
from typing import Protocol, Literal, Any
from pydantic import Field, BaseModel


EFFORT = Literal["none", "low", "medium", "high", "xhigh", "max"]
TOOL_CHOICE = Literal["auto", "required", "none"]

@dataclass(frozen=True)
class Message:
    role: Literal["tool", "assistant", "system", "user"]
    content: str


class ProviderCapabilities(BaseModel):
    supports_temperature: bool = False
    supports_reasoning_effort: bool = False
    supports_system_messages: bool = True
    supports_structured_output_native: bool = False
    supports_tools: bool = False
    supports_parallel_tool_calls: bool = False
    supports_streaming: bool = False
    allowed_tool_choice_modes: frozenset[TOOL_CHOICE] = frozenset({"auto", "none"})


class ToolSpec(BaseModel):
    name: str
    description: str
    parameters: dict[str, Any]


class LLMRequest(BaseModel):
    model: str
    messages: list[Message]
    temperature: float | None = Field(default=None, ge=0.0, le=2.0)
    max_tokens: int | None = Field(default=None, gt=0)
    top_p: float | None = Field(default=None, gt=0.0, le=1.0)
    top_k: int | None = Field(default=None, ge=1)
    frequency_penalty: float | None = Field(default=0, ge=-2.0, le=2.0)
    tools: list[ToolSpec] | None = None
    tool_choice: TOOL_CHOICE | dict[str, Any] | None = None
    effort: EFFORT | None = None
    timeout_s: float | None = Field(default=None, gt=0)
    seed: int | None = None
    response_schema: dict[str, Any] | type[BaseModel] | None = None


class ToolCall(BaseModel):
    id: str
    name: str
    args: dict[str, Any]

class LLMResponsePayload(BaseModel):
    text: str | None
    structure: dict[str, Any] | None = None
    tool_calls: list[ToolCall] = Field(default_factory=list)

class LLMResponseLineage(BaseModel):
    provider: str
    model: str
    request_id: str
    latency_ms: int

class Usage(BaseModel):
    input_tokens: int
    output_tokens: int
    cached_tokens: int = 0
    reasoning_tokens: int = 0
    cost_usd: float

class LLMResponseExecution(BaseModel):
    stop_reason: str
    safety_triggered: bool



class LLMResponse(BaseModel):
    payload: LLMResponsePayload
    lineage: LLMResponseLineage
    ledger: Usage
    execution: LLMResponseExecution

class LLMError(Exception):
    def __init__(
        self, 
        message: str, 
        provider: str | None = None, 
        raw_error: Exception | None = None
    ):
        super().__init__(message)
        self.provider = provider
        self.raw_error = raw_error

class LLMTimeout(LLMError):
    pass

class LLMBudgetExceeded(LLMError):
    pass

class LLMInvalidOutput(LLMError):
    pass
class LLMProvider(Protocol):
    @property
    def capabilities(self) -> ProviderCapabilities:
        ...


    async def generate(
        self,
        request: LLMRequest,
    ) -> LLMResponse:
        ...
