CREATE DATABASE TTTMatches;

\c TTTMatches;

CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    movementsCode INTEGER,   
    winner CHAR,
    CONSTRAINT valid_winner CHECK (winner IN ('X', 'O', 'D'))
);