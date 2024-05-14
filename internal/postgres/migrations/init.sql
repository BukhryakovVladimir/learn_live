DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'learn_live') THEN
        CREATE DATABASE learn_live;
END IF;
END $$;

\c learn_live;

-- Create schema
CREATE SCHEMA IF NOT EXISTS public;

CREATE TABLE IF NOT EXISTS group_uni (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    group_name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS person (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(60) NOT NULL,
    firstname VARCHAR(255),
    lastname VARCHAR(255),
    email VARCHAR(255),
    phone_number VARCHAR(20),
    group_id INTEGER,
    is_professor BOOL NOT NULL,
    is_admin BOOL NOT NULL,
    sex VARCHAR(10),
    birthdate DATE,
    FOREIGN KEY (group_id) REFERENCES group_uni(id)
);

CREATE TABLE IF NOT EXISTS subject (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    subject_name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS room (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    subject_id INTEGER,
    room_name VARCHAR(255),
    FOREIGN KEY (subject_id) REFERENCES subject(id),
    UNIQUE (subject_id, room_name)
);

CREATE TABLE IF NOT EXISTS group_subject (
    group_id INTEGER,
    subject_id INTEGER,
    FOREIGN KEY (group_id) REFERENCES group_uni(id),
    FOREIGN KEY (subject_id) REFERENCES subject(id),
    PRIMARY KEY (group_id, subject_id)
);

CREATE TABLE IF NOT EXISTS professor_subject (
    professor_id INTEGER,
    subject_id INTEGER,
    FOREIGN KEY (professor_id) REFERENCES person(id),
    FOREIGN KEY (subject_id) REFERENCES subject(id),
    PRIMARY KEY (professor_id, subject_id)
);

CREATE TABLE IF NOT EXISTS professor_group (
    professor_id INTEGER,
    group_id INTEGER,
    FOREIGN KEY (professor_id) REFERENCES person(id),
    FOREIGN KEY (group_id) REFERENCES group_uni(id),
    PRIMARY KEY (professor_id, group_id)
);

CREATE TABLE IF NOT EXISTS student_grades (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    student_id INTEGER,
    subject_id INTEGER,
    grade INTEGER DEFAULT 0,
    has_attended BOOL DEFAULT true,
    FOREIGN KEY (student_id) REFERENCES person(id),
    FOREIGN KEY (subject_id) REFERENCES subject(id)
);

CREATE TABLE IF NOT EXISTS student_total_grades (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    student_id INTEGER,
    subject_id INTEGER,
    grade VARCHAR(50),
    FOREIGN KEY (student_id) REFERENCES person(id),
    FOREIGN KEY (subject_id) REFERENCES subject(id),
    UNIQUE (student_id, subject_id)
);