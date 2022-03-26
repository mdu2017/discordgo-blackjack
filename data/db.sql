CREATE DATABASE dgblackjack
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;


INSERT INTO Player(user_id, guild_id, username, credits, wins, losses, user_rank)
VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING

CREATE TABLE IF NOT EXISTS Player(
    user_id   varchar(20),
    guild_id  varchar(20),
    username  varchar(70),
    credits   int,
    wins      int,
    losses    int,
    user_rank int,

    PRIMARY KEY(user_id, guild_id)
    );