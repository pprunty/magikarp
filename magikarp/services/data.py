"""Service class for handling the ingestion of Dan's device's data."""
import pandas as pd
import json
import os
import logging
from typing import Any, Dict, List

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

class DataService:
    """Service for loading and managing user data from various sources."""

    def __init__(self, base_path: str = 'data'):
        """Initializes the DataService with the base data directory.

        Args:
            base_path (str): The base directory where data files are located.
        """
        self.base_path = os.path.join(os.getcwd(), base_path)
        self._validate_data_directory()
        self._calendar_data = None
        self._location_data = None
        self._social_media_data = None
        self._spotify_playlists_data = None
        self._user_profile_data = None

    def _validate_data_directory(self):
        """Validates that the data directory exists."""
        if not os.path.exists(self.base_path):
            logger.error(f"Data directory {self.base_path} does not exist.")
            raise FileNotFoundError(f"Data directory {self.base_path} does not exist.")
        logger.info(f"Data directory {self.base_path} found.")

    def _load_json(self, filepath: str) -> Dict:
        """Loads a JSON file from the given filepath.

        Args:
            filepath (str): The path to the JSON file.

        Returns:
            Dict: The parsed JSON data.

        Raises:
            FileNotFoundError: If the file does not exist.
            json.JSONDecodeError: If the file is not valid JSON.
        """
        try:
            with open(filepath, 'r') as file:
                logger.info(f"Loading JSON data from {filepath}")
                return json.load(file)
        except (FileNotFoundError, json.JSONDecodeError) as e:
            logger.error(f"Error loading JSON data from {filepath}: {e}")
            raise

    @property
    def calendar_data(self) -> pd.DataFrame:
        """Loads and returns the calendar data.

        Returns:
            pd.DataFrame: The calendar data.
        """
        if self._calendar_data is None:
            logger.info("Loading calendar data")
            self._calendar_data = pd.read_csv(os.path.join(self.base_path, 'calendar.csv'))
        return self._calendar_data

    @property
    def location_data(self) -> pd.DataFrame:
        """Loads and returns the location data.

        Returns:
            pd.DataFrame: The location data.
        """
        if self._location_data is None:
            logger.info("Loading location data")
            self._location_data = pd.read_csv(os.path.join(self.base_path, 'location.csv'))
        return self._location_data

    @property
    def social_media_data(self) -> Dict:
        """Loads and returns the social media data.

        Returns:
            Dict: The social media data.
        """
        if self._social_media_data is None:
            self._social_media_data = self._load_json(os.path.join(self.base_path, 'social_media.json'))
        return self._social_media_data

    @property
    def spotify_playlists_data(self) -> Dict:
        """Loads and returns the Spotify playlists data.

        Returns:
            Dict: The Spotify playlists data.
        """
        if self._spotify_playlists_data is None:
            self._spotify_playlists_data = self._load_json(os.path.join(self.base_path, 'spotify_playlists.json'))
        return self._spotify_playlists_data

    @property
    def user_profile_data(self) -> Dict:
        """Loads and returns the user profile data.

        Returns:
            Dict: The user profile data.
        """
        if self._user_profile_data is None:
            self._user_profile_data = self._load_json(os.path.join(self.base_path, 'user_profile.json'))
        return self._user_profile_data

    def get_formatted_data(self) -> str:
        """Formats and returns the user data for output.

        Returns:
            str: The formatted user data.
        """
        logger.info("Formatting data for output")
        recent_posts: List[str] = self.social_media_data.get('twitter', {}).get('recent_posts', [])
        fitness_data: Dict[str, Any] = self.user_profile_data.get('fitness_data', {})
        playlists: List[Dict[str, Any]] = self.spotify_playlists_data.get('playlists', [])
        calendar: str = self.calendar_data.to_string(index=False)

        formatted_data: str = f"""
        User Metadata
        ---
        Recent Social Media Posts: {', '.join(recent_posts)}
        Fitness Data: {fitness_data}
        Music Playlists: {playlists}
        Calendar: {calendar}
        """

        # Add more keys as appropriate, such as location and notifications
        location: Dict[str, Any] = self.user_profile_data.get('location', {})
        notifications: List[Dict[str, Any]] = self.user_profile_data.get('previous_notifications', [])
        app_usage: Dict[str, Any] = self.user_profile_data.get('app_usage', {})

        formatted_data += f"""
        Location: {location}
        Notifications: {notifications}
        App Usage: {app_usage}
        ---
        \n\n
        """

        return formatted_data
