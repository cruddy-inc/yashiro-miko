from openai import AsyncOpenAI
from miko.llm.base import LLMProvider, Message
import os
from miko.core.logging.logger import get_logger
from miko.core.logging.errors import UnauthorizedError
from miko.core.settings import Settings

settings = Settings()
okitsune = get_logger("openai")


class OpenAIProvider(LLMProvider):
    def __init__(
        self,
        model: str,
        api_key: str | None = None,
    ) -> None:

        self.model = model

        settings_key = (
            settings.OPENAI_API_KEY.get_secret_value()
            if settings.OPENAI_API_KEY
            else None
        )

        resolved_key = api_key or os.getenv("OPENAI_API_KEY") or settings_key
        if not resolved_key or not resolved_key.strip():
            okitsune.error(
                "Missing OpenAI API key. Set OPENAI_API_KEY or pass api_key to the client."
            )
            raise UnauthorizedError(
                message="The api_key client option must be set either by passing api_key to the client or by setting the OPENAI_API_KEY environment variable.",
                provider="openai",
                env_var="OPENAI_API_KEY",
            )

        self._client = AsyncOpenAI(api_key=resolved_key.strip())

    async def generate(
        self,
        messages: list[Message],
    ) -> str:

        response = await self._client.chat.completions.create(
            model=self.model,
            messages=[{"role": msg.role, "content": msg.content} for msg in messages],
        )
        return response.choices[0].message.content or ""
