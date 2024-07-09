import pandas as pd
import json
import os


class DataService:
    def __init__(self, **kwargs):
        self.base_path = os.path.join(os.getcwd(), 'data')
        self.calendar_data = pd.read_csv(os.path.join(self.base_path, 'calendar.csv'))
        self.location_data = pd.read_csv(os.path.join(self.base_path, 'location.csv'))
        self.social_media_data = self._load_json(os.path.join(self.base_path, 'social_media.json'))
        self.spotify_playlists_data = self._load_json(os.path.join(self.base_path, 'spotify_playlists.json'))
        self.user_profile_data = self._load_json(os.path.join(self.base_path, 'user_profile.json'))

    def _load_json(self, filepath):
        with open(filepath, 'r') as file:
            return json.load(file)

    def get_calendar_data(self):
        return self.calendar_data

    def get_location_data(self):
        return self.location_data

    def get_social_media_data(self):
        return self.social_media_data

    def get_spotify_playlists_data(self):
        return self.spotify_playlists_data

    def get_user_profile_data(self):
        return self.user_profile_data

    def get_formatted_data(self):
        recent_posts = self.social_media_data.get('twitter', {}).get('recent_posts', [])
        fitness_data = self.user_profile_data.get('fitness_data', {})
        playlists = self.spotify_playlists_data.get('playlists', [])
        calendar = self.calendar_data.to_string(index=False)

        formatted_data = f"""
        User Metadata
        ---
        Recent Social Media Posts: {', '.join(recent_posts)}
        Fitness Data: {fitness_data}
        Music Playlists: {playlists}
        Calendar: {calendar}
        """

        # Add more keys as appropriate, such as location and notifications
        location = self.user_profile_data.get('location', {})
        notifications = self.user_profile_data.get('previous_notifications', [])
        app_usage = self.user_profile_data.get('app_usage', {})

        formatted_data += f"""
        Location: {location}
        Notifications: {notifications}
        App Usage: {app_usage}
        ---
        \n\n
        """

        return formatted_data
