from abc import ABC, abstractmethod


class PluginBase(ABC):
    """Base class for all plugins."""

    @abstractmethod
    def process_prompt(self, prompt: str) -> str:
        """
        Process the prompt before it's sent to the model.

        Args:
            prompt (str): The original prompt

        Returns:
            str: The modified prompt
        """
        pass

    @property
    @abstractmethod
    def name(self) -> str:
        """Return the name of the plugin."""
        pass
