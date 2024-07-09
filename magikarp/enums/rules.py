from enum import Enum, unique


@unique
class RuleSetEnum(str, Enum):
    """Enumeration of possible rules for push notifications."""
    no_sleep_notifications = "Notifications should not take place during user’s inferred sleeping hours."
    meeting_reminders = "User push notifications for meeting reminders should arrive 30 minutes before the meeting's start time."
    music_notifications = "Notifications should promptly suggest new recommended music based on user playlist data"
    health_notifications = "Notifications should prompt user to take more steps if below average or to go to sleep to ensure average sleep goal is met"
    entertainment_notifications_timing = "Entertainment notifications should typically only take place before and after user’s inferred working hours."
    short_notifications = "Push notifications should be no more than 2 sentences."
    friendly_tone = "The tone in push notifications should always be friendly."
    encourage_social_interactions = "Push notifications should encourage social interactions with contacts stored on mobile device."
    warn_unhealthy_behaviors = "Push notifications should warn user of unhealthy behaviors."