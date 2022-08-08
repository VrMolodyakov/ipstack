CREATE TABLE users(
    user_id SERIAL PRIMARY KEY NOT NULL,
    nickname VARCHAR(100) NOT NULL
);

CREATE TABLE ip_info(
    ip_id SERIAL PRIMARY KEY,
    ip VARCHAR(48),
    continent_name VARCHAR(255) NOT NULL,
    country_name VARCHAR(255),
    region_name VARCHAR(255),
    city VARCHAR(255),
    zip VARCHAR(255),
    latitude DECIMAL,
    longitude DECIMAL
);

CREATE TABLE user_ip_info(
    user_ip_info_id SERIAL PRIMARY KEY,
    user_id int references users(user_id) ON DELETE CASCADE,
    ip_id int references ip_info(ip_id) ON DELETE CASCADE
);

