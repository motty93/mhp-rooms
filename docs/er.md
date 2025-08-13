# ER図（Entity Relationship Diagram）

## 概要
MonHubのデータベース構造を示すER図です。

## テーブル関係図

```mermaid
erDiagram
    users ||--o{ rooms : "hosts"
    users ||--o{ room_members : "joins"
    users ||--o{ room_messages : "sends"
    users ||--o{ user_blocks : "blocker"
    users ||--o{ user_blocks : "blocked"
    users ||--o{ player_names : "has"
    users ||--o{ password_resets : "requests"
    users ||--o{ user_follows : "follower"
    users ||--o{ user_follows : "following"
    
    game_versions ||--o{ rooms : "for"
    game_versions ||--o{ player_names : "for"
    
    rooms ||--o{ room_members : "has"
    rooms ||--o{ room_messages : "contains"
    rooms ||--o{ room_logs : "logs"

    users {
        uuid id PK
        uuid supabase_user_id UK
        string email UK
        string username UK "NULL可"
        string display_name
        string avatar_url "NULL可"
        string bio "NULL可"
        string psn_online_id "NULL可"
        string nintendo_network_id "NULL可"
        string nintendo_switch_id "NULL可"
        string pretendo_network_id "NULL可"
        string twitter_id "NULL可"
        jsonb favorite_games "お気に入りゲーム"
        jsonb play_times "プレイ時間帯"
        boolean is_active
        string role
        timestamp created_at
        timestamp updated_at
    }

    game_versions {
        uuid id PK
        string code UK
        string name
        string short_name
        string platform "追加予定"
        integer display_order
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    rooms {
        uuid id PK
        string room_code UK
        string name
        string description "NULL可"
        uuid game_version_id FK
        uuid host_user_id FK
        integer max_players
        string password_hash "NULL可"
        string target_monster "NULL可"
        string rank_requirement "NULL可"
        boolean is_active
        boolean is_closed
        timestamp created_at
        timestamp updated_at
    }

    room_members {
        uuid id PK
        uuid room_id FK
        uuid user_id FK
        integer player_number
        boolean is_host
        timestamp joined_at
        timestamp left_at "NULL可"
    }

    room_messages {
        uuid id PK
        uuid room_id FK
        uuid user_id FK
        string message_type
        text content
        boolean is_deleted
        timestamp created_at
        timestamp updated_at
    }

    room_logs {
        uuid id PK
        uuid room_id FK
        uuid user_id FK "NULL可"
        string action_type
        jsonb details
        timestamp created_at
    }

    user_blocks {
        uuid id PK
        uuid blocker_user_id FK
        uuid blocked_user_id FK
        string reason "NULL可"
        timestamp created_at
    }

    player_names {
        uuid id PK
        uuid user_id FK
        uuid game_version_id FK
        string name
        timestamp created_at
        timestamp updated_at
    }

    password_resets {
        uuid id PK
        uuid user_id FK
        string token UK
        timestamp expires_at
        boolean used
        timestamp created_at
        timestamp updated_at
    }

    user_follows {
        uuid id PK
        uuid follower_user_id FK
        uuid following_user_id FK
        string status "pending/accepted"
        timestamp created_at
        timestamp updated_at
    }
```

## インデックス

### users
- `idx_users_supabase_user_id` (supabase_user_id)
- `idx_users_email` (email)
- `idx_users_username` (username)
- `idx_users_is_active` (is_active)

### game_versions
- `idx_game_versions_code` (code)
- `idx_game_versions_is_active` (is_active)
- `idx_game_versions_display_order` (display_order)

### rooms
- `idx_rooms_room_code` (room_code)
- `idx_rooms_game_version_id` (game_version_id)
- `idx_rooms_host_user_id` (host_user_id)
- `idx_rooms_is_active` (is_active)
- `idx_rooms_is_closed` (is_closed)
- `idx_rooms_created_at` (created_at)

### room_members
- `idx_room_members_room_id` (room_id)
- `idx_room_members_user_id` (user_id)
- `idx_room_members_room_user` (room_id, user_id)
- `idx_room_members_left_at` (left_at)

### room_messages
- `idx_room_messages_room_id` (room_id)
- `idx_room_messages_user_id` (user_id)
- `idx_room_messages_created_at` (created_at)
- `idx_room_messages_is_deleted` (is_deleted)

### room_logs
- `idx_room_logs_room_id` (room_id)
- `idx_room_logs_user_id` (user_id)
- `idx_room_logs_action_type` (action_type)
- `idx_room_logs_created_at` (created_at)

### user_blocks
- `idx_user_blocks_blocker_user_id` (blocker_user_id)
- `idx_user_blocks_blocked_user_id` (blocked_user_id)
- `idx_user_blocks_blocker_blocked` (blocker_user_id, blocked_user_id)

### player_names
- `idx_player_names_user_game` (user_id, game_version_id)

### password_resets
- `idx_password_resets_user_id` (user_id)
- `idx_password_resets_token` (token)
- `idx_password_resets_expires_at` (expires_at)

### user_follows
- `idx_user_follows_follower_user_id` (follower_user_id)
- `idx_user_follows_following_user_id` (following_user_id)
- `idx_user_follows_follower_following` (follower_user_id, following_user_id)
- `idx_user_follows_status` (status)