"""Hello agent built on Google Agent Development Kit."""

from google.adk import Agent

DEFAULT_MODEL = "gemini-2.5-flash"
DEFAULT_NAME = "hello_agent"
DEFAULT_INSTRUCTION = (
    "You are a helpful fintech playground assistant. "
    "Greet the user and offer to help with questions about this repo."
)


def build_hello_agent(
    name: str = DEFAULT_NAME,
    model: str = DEFAULT_MODEL,
    instruction: str = DEFAULT_INSTRUCTION,
) -> Agent:
    return Agent(name=name, model=model, instruction=instruction)


root_agent = build_hello_agent()
