-- =====================================================================
--  YMMO — Demo seed data
--  Loaded automatically by docker compose on first startup
--  (mounted as /docker-entrypoint-initdb.d/03-seed.sql, after the schema).
--  Also runnable by hand:  mariadb -u root -p ymmo < seed.sql
--  Safe to re-run: it clears these tables first.
--
--  Test accounts (password in clear here ONLY because this is demo data):
--    agent@ymmo.fr    / agent1234     (role AGENT)
--    director@ymmo.fr / director1234  (role DIRECTOR)
--    buyer@ymmo.fr    / buyer1234     (role BUYER)
-- =====================================================================
USE ymmo;
SET NAMES utf8mb4;  -- preserve accented names (e.g. "Hélène")

-- Reset all demo data. FK checks are disabled so the reset is re-runnable even
-- after data has been created through the app (conversations, sales, etc.).
SET FOREIGN_KEY_CHECKS = 0;
DELETE FROM messages;
DELETE FROM conversations;
DELETE FROM visits;
DELETE FROM meeting_participants;
DELETE FROM meetings;
DELETE FROM sale_files;
DELETE FROM property_views;
DELETE FROM favorites;
DELETE FROM alerts;
DELETE FROM property_photos;
DELETE FROM properties;
DELETE FROM users;            -- full reset (avoids id collisions with app sign-ups)
DELETE FROM agencies;
SET FOREIGN_KEY_CHECKS = 1;

-- ---------------------------------------------------------------------
-- Agencies (1 = HQ Aix-en-Provence, then a few regional agencies)
-- ---------------------------------------------------------------------
INSERT INTO agencies (id, name, city, postal_code, is_hq) VALUES
  (1, 'Ymmo Siège',  'Aix-en-Provence', '13100', TRUE),
  (2, 'Ymmo Paris',  'Paris',           '75001', FALSE),
  (3, 'Ymmo Lyon',   'Lyon',            '69002', FALSE),
  (4, 'Ymmo Marseille','Marseille',     '13001', FALSE);

-- ---------------------------------------------------------------------
-- Users (bcrypt hashes, cost 10 — verify in Go's bcrypt)
-- role_id:  2=BUYER, 4=AGENT, 5=DIRECTOR   department_id: 1=Management, 2=Sales
-- ---------------------------------------------------------------------
INSERT INTO users (id, email, password_hash, last_name, first_name, role_id, department_id, agency_id, is_active) VALUES
  (1, 'agent@ymmo.fr',    '$2b$10$TVzS9UGprQwiyuN6QlTUTudailOAbUZ.cyNfYaCnZ3XydYkQoK5Qa', 'Martin',  'Alice', 4, 2, 1, TRUE),
  (2, 'director@ymmo.fr', '$2b$10$K0Mm7jiJTYRDjv5xXNmb3O.TCRROlns0P7UmpwOFWD0j7JO54ZSmC', 'Bernard', 'Marc',  5, 1, 1, TRUE),
  (3, 'buyer@ymmo.fr',    '$2b$10$eBp33kzUuDCIwV3YG5tiz.pzFFLl.P8AjjgaAQb0jqQHZBXPZS31K', 'Petit',   'Sophie', 2, NULL, NULL, TRUE),
  (4, 'hq@ymmo.fr',       '$2b$10$byR4yyTCwTVQvCPA/KP2TuoPC6Y1oCm87LjjAT5gkzF6YPtEeMFIe', 'Durand',  'Hélène', 6, 1, 1, TRUE),
  (5, 'it@ymmo.fr',       '$2b$10$1A8gzlv.ewQPw0ud/CJz0.541JL2ASUEyvoeKBrY2UNxIQDuNoc2e', 'Rousseau','Karim',  7, 5, 1, TRUE),
  (6, 'director.paris@ymmo.fr', '$2b$10$I1SvMoSDb37EDcrs/MnbcONTNTIkgyzaVlBvrDUJWLCX/6U/vp7Wa', 'Moreau', 'Julie', 5, 1, 2, TRUE);

-- ---------------------------------------------------------------------
-- Properties
-- category_id: 1=House 2=Apartment 3=Studio 5=Office
-- 4 are AVAILABLE (public), 1 is DRAFT (must stay hidden from the catalogue)
-- ---------------------------------------------------------------------
INSERT INTO properties
  (reference, title, description, category_id, status, price, area, rooms, bedrooms,
   energy_rating, city, postal_code, is_exclusive, agency_id, agent_id, published_at)
VALUES
  ('SEED-0001', 'Appartement lumineux T3', 'Proche centre, balcon sud', 2, 'AVAILABLE', 480000, 75, 3, 2, 'C', 'Paris',           '75011', TRUE,  2, 1, NOW()),
  ('SEED-0002', 'Maison familiale 5 pièces', 'Jardin 400m², garage',    1, 'AVAILABLE', 650000, 140, 5, 4, 'D', 'Aix-en-Provence', '13100', FALSE, 1, 1, NOW()),
  ('SEED-0003', 'Studio étudiant',          'Idéal investissement',     3, 'AVAILABLE', 180000, 28,  1, 0, 'E', 'Lyon',            '69002', FALSE, 3, 1, NOW()),
  ('SEED-0004', 'Bureau open-space',        'Quartier d''affaires',     5, 'AVAILABLE', 320000, 90,  4, 0, 'C', 'Marseille',       '13001', FALSE, 4, 1, NOW()),
  ('SEED-0005', 'Appartement à valider',    'En attente de publication',2, 'DRAFT',     550000, 82,  3, 2, 'B', 'Paris',           '75008', TRUE,  2, 1, NULL),
  ('SEED-0006', 'Appartement Haussmannien', 'Charme parisien, moulures', 2, 'AVAILABLE', 720000, 95,  4, 2, 'C', 'Paris',           '75009', FALSE, 2, 6, NOW()),
  ('SEED-0007', 'Studio République',        'Idéal investissement locatif',3,'AVAILABLE',245000, 30,  1, 0, 'D', 'Paris',           '75011', FALSE, 2, 6, NOW());

-- A couple of photos on the first property.
INSERT INTO property_photos (property_id, url, sort_order, is_primary)
SELECT id, 'https://placehold.co/800x600?text=Photo+1', 0, TRUE  FROM properties WHERE reference='SEED-0001';
INSERT INTO property_photos (property_id, url, sort_order, is_primary)
SELECT id, 'https://placehold.co/800x600?text=Photo+2', 1, FALSE FROM properties WHERE reference='SEED-0001';
