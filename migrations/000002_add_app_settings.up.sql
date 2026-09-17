CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY, 
    value JSONB NOT NULL, 
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_settings (key, value) 
VALUES ('app_name', 'Studio KAIFA Bot');

INSERT INTO app_settings (key, value) 
VALUES ('opening_hours', '{
    "monday": {"start_time": "10:00", "end_time": "22:00"},
    "tuesday": {"start_time": "10:00", "end_time": "22:00"},
    "wednesday": {"start_time": "10:00", "end_time": "22:00"},
    "thursday": {"start_time": "10:00", "end_time": "22:00"},
    "friday": {"start_time": "10:00", "end_time": "22:00"},
    "saturday": {"start_time": "10:00", "end_time": "22:00"},
    "sunday": {"start_time": "10:00", "end_time": "22:00"}
}');

INSERT INTO app_settings (key, value) 
VALUES ('custom_prices', '[
    {
        "type": "base",
        "days_of_week": "monday,tuesday,wednesday,thursday,friday",
        "start_time": "10:00",
        "end_time": "22:00",
        "price": 700,
        "name": "Base",
        "description": "Base price for the regular hours."
    },
    {
        "type": "base",
        "days_of_week": "saturday,sunday",
        "start_time": "10:00",
        "end_time": "22:00",
        "price": 1000,
        "name": "Base",
        "description": "Base price for the weekend hours."
    },
    {
        "type": "early_bird",
        "days_of_week": "monday,tuesday,wednesday,thursday,friday",
        "start_time": "10:00",
        "end_time": "14:00",
        "price": 800,
        "name": "Early Bird",
        "description": "Early Bird is a special price for the early morning hours."
    },
    {
        "type": "regular",
        "days_of_week": "monday,tuesday,wednesday,thursday,friday",
        "start_time": "14:00",
        "end_time": "22:00",
        "price": 1000,
        "name": "Regular",
        "description": "Regular is a special price for the regular hours."
    },
    {
        "type": "weekend_early_bird",
        "days_of_week": "saturday,sunday",
        "start_time": "10:00",
        "end_time": "14:00",
        "price": 1000,
        "name": "Weekend Early Bird",
        "description": "Weekend is a special price for the weekend hours."
    },
    {
        "type": "weekend",
        "days_of_week": "saturday,sunday",
        "start_time": "14:00",
        "end_time": "22:00",
        "price": 1200,
        "name": "Weekend",
        "description": "Weekend is a special price for the weekend hours."
    }
]');

INSERT INTO app_settings (key, value) 
VALUES ('bot_admins', '[
    {
        "telegram_id": 329210636,
        "name": "Даня Марля",
        "phone": "+7 913 245 1578",
        "username": "@mvrlya"
        "roles": ["admin", "owner", "producer"]
    },
    {
        "telegram_id": 279655008,
        "name": "Victor Nagaev",
        "phone": "",
        "username": "@victorfuckingnagaev",
        "roles": ["teacher", "producer"]
    },
    {
        "telegram_id": 338074569,
        "name": "Artur ROOS",
        "phone": "+7 965 057 7800",
        "username": "@ArthurRooss",
        "roles": ["teacher", "admin", "owner", "producer"]
    },
    {
        "telegram_id": 576077782,
        "name": "Andrey AMADAU",
        "phone": "+7 921 402 3791",
        "username": "@amadau",
        "roles": ["teacher", "admin", "owner", "producer"]
    },
    {
        "telegram_id": 0,
        "name": "Феликс",
        "phone": "+7 981 257 4213",
        "username": "",
        "roles": ["owner", "producer"]
    },
]');

