CREATE USER mainflux;
CREATE DATABASE users;
CREATE DATABASE clients;
GRANT ALL PRIVILEGES ON DATABASE users TO mainflux;
GRANT ALL PRIVILEGES ON DATABASE clients TO mainflux;
