CREATE DATABASE TTTMatches;

\c TTTMatches;

CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    movementsCode INTEGER,   
    winner CHAR,
    CONSTRAINT valid_winner CHECK (winner IN ('X', 'O', 'D'))
);

CREATE TABLE movements (
    id SERIAL PRIMARY KEY,
    movementNumber INTEGER CHECK (movementNumber >= 1 AND movementNumber <= 9),
    isWinner BOOLEAN,
    stateCode INTEGER,
    match_id INTEGER REFERENCES matches(id)
);