-- =====================================================================
--  YMMO — MariaDB relational schema (Database module)
--  Normalized model (3NF) · InnoDB engine · utf8mb4 encoding
--  Run on MariaDB 10.6+ :  mariadb -u root -p < schema.sql
--  (older MariaDB installs may still expose the legacy `mysql` client)
-- =====================================================================

DROP DATABASE IF EXISTS ymmo;
CREATE DATABASE ymmo
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
USE ymmo;

-- ---------------------------------------------------------------------
-- 1. LOOKUP TABLES
--    Fixed values are stored in dedicated tables instead of being
--    repeated as strings -> normalization + consistency.
-- ---------------------------------------------------------------------

-- Internal company departments (the "matrice des droits" from the brief)
CREATE TABLE departments (
  id          TINYINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(50) NOT NULL UNIQUE,   -- Management, Sales, Comm. & Mktg, Admin/HR, IT & Support
  description VARCHAR(255) NULL
) ENGINE=InnoDB;

-- Application roles (RBAC)
CREATE TABLE roles (
  id          TINYINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code        VARCHAR(30) NOT NULL UNIQUE,   -- VISITOR, BUYER, SELLER, AGENT, DIRECTOR, HQ, IT
  label       VARCHAR(80) NOT NULL,
  is_internal BOOLEAN NOT NULL DEFAULT FALSE -- TRUE = Ymmo employee
) ENGINE=InnoDB;

-- Property categories (residential / commercial -> sub-types)
CREATE TABLE property_categories (
  id      TINYINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  sector  ENUM('RESIDENTIAL','COMMERCIAL') NOT NULL,
  label   VARCHAR(50) NOT NULL UNIQUE        -- House, Apartment, Studio, Office, Retail space...
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 2. AGENCIES
-- ---------------------------------------------------------------------
CREATE TABLE agencies (
  id          SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(120) NOT NULL,
  city        VARCHAR(100) NOT NULL,
  address     VARCHAR(255) NULL,
  postal_code VARCHAR(10)  NULL,
  phone       VARCHAR(20)  NULL,
  email       VARCHAR(150) NULL,
  is_hq       BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE for the Aix-en-Provence head office
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_agencies_city (city)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 3. USERS
--    A single users table for every actor, distinguished by role_id.
--    department_id and agency_id only apply to internal users
--    -> nullable (a buyer has neither department nor agency).
-- ---------------------------------------------------------------------
CREATE TABLE users (
  id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email         VARCHAR(150) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,        -- bcrypt, never plaintext
  last_name     VARCHAR(80)  NOT NULL,
  first_name    VARCHAR(80)  NOT NULL,
  phone         VARCHAR(20)  NULL,
  role_id       TINYINT UNSIGNED NOT NULL,
  department_id TINYINT UNSIGNED NULL,        -- internal only
  agency_id     SMALLINT UNSIGNED NULL,       -- internal only
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_users_role       FOREIGN KEY (role_id)       REFERENCES roles(id),
  CONSTRAINT fk_users_department FOREIGN KEY (department_id) REFERENCES departments(id),
  CONSTRAINT fk_users_agency     FOREIGN KEY (agency_id)     REFERENCES agencies(id),
  INDEX idx_users_role (role_id),
  INDEX idx_users_agency (agency_id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 4. ACCESS MATRIX (department -> department)
--    Direct translation of the brief's table. Lets the app (and the
--    INFRA documentation) answer: "what right does department X have
--    over department Y's files?".
-- ---------------------------------------------------------------------
CREATE TABLE department_permissions (
  requester_department_id TINYINT UNSIGNED NOT NULL, -- who accesses
  target_department_id    TINYINT UNSIGNED NOT NULL, -- which department's files
  access_level            ENUM('NONE','READ','READ_WRITE') NOT NULL,
  PRIMARY KEY (requester_department_id, target_department_id),
  CONSTRAINT fk_perm_requester FOREIGN KEY (requester_department_id) REFERENCES departments(id),
  CONSTRAINT fk_perm_target    FOREIGN KEY (target_department_id)    REFERENCES departments(id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 5. PROPERTIES (core entity)
-- ---------------------------------------------------------------------
CREATE TABLE properties (
  id               INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  reference        VARCHAR(20) NOT NULL UNIQUE,        -- e.g. YMMO-2026-00042
  title            VARCHAR(150) NOT NULL,
  description      TEXT NULL,
  category_id      TINYINT UNSIGNED NOT NULL,
  status           ENUM('DRAFT','PENDING_REVIEW','AVAILABLE',
                        'UNDER_OFFER','SOLD','WITHDRAWN') NOT NULL DEFAULT 'DRAFT',
  price            DECIMAL(12,2) NOT NULL,
  area             DECIMAL(8,2)  NOT NULL,             -- m²
  rooms            TINYINT UNSIGNED NULL,
  bedrooms         TINYINT UNSIGNED NULL,
  bathrooms        TINYINT UNSIGNED NULL,
  floor            TINYINT NULL,
  build_year       SMALLINT UNSIGNED NULL,
  energy_rating    ENUM('A','B','C','D','E','F','G') NULL,  -- DPE
  ghg_rating       ENUM('A','B','C','D','E','F','G') NULL,  -- GES
  address          VARCHAR(255) NULL,
  city             VARCHAR(100) NOT NULL,
  postal_code      VARCHAR(10)  NOT NULL,
  latitude         DECIMAL(10,7) NULL,
  longitude        DECIMAL(10,7) NULL,
  is_exclusive     BOOLEAN NOT NULL DEFAULT FALSE,     -- "Exclusivité" badge (style guide)
  view_count       INT UNSIGNED NOT NULL DEFAULT 0,    -- fast counter (popular properties)
  agency_id        SMALLINT UNSIGNED NOT NULL,
  agent_id         INT UNSIGNED NULL,                  -- responsible agent
  seller_id        INT UNSIGNED NULL,                  -- originating seller (client)
  approved_by      INT UNSIGNED NULL,                  -- agent who approved the listing
  approved_at      TIMESTAMP NULL,
  published_at     TIMESTAMP NULL,
  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_properties_category FOREIGN KEY (category_id) REFERENCES property_categories(id),
  CONSTRAINT fk_properties_agency   FOREIGN KEY (agency_id)   REFERENCES agencies(id),
  CONSTRAINT fk_properties_agent    FOREIGN KEY (agent_id)    REFERENCES users(id),
  CONSTRAINT fk_properties_seller   FOREIGN KEY (seller_id)   REFERENCES users(id),
  CONSTRAINT fk_properties_approver FOREIGN KEY (approved_by) REFERENCES users(id),
  CONSTRAINT chk_properties_price   CHECK (price >= 0),
  CONSTRAINT chk_properties_area    CHECK (area > 0),
  -- Indexes designed for the search engine (price/city/area/energy filters)
  INDEX idx_properties_status (status),
  INDEX idx_properties_city (city),
  INDEX idx_properties_price (price),
  INDEX idx_properties_area (area),
  INDEX idx_properties_category (category_id),
  INDEX idx_properties_agent (agent_id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 6. PROPERTY PHOTOS (1 property -> N photos)
-- ---------------------------------------------------------------------
CREATE TABLE property_photos (
  id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  property_id INT UNSIGNED NOT NULL,
  url         VARCHAR(255) NOT NULL,
  sort_order  TINYINT UNSIGNED NOT NULL DEFAULT 0,
  is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT fk_photos_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE,
  INDEX idx_photos_property (property_id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 7. FAVORITES (N-N relation user <-> property)
-- ---------------------------------------------------------------------
CREATE TABLE favorites (
  user_id     INT UNSIGNED NOT NULL,
  property_id INT UNSIGNED NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, property_id),
  CONSTRAINT fk_fav_user     FOREIGN KEY (user_id)     REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_fav_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 8. ALERTS (saved search criteria for a buyer)
-- ---------------------------------------------------------------------
CREATE TABLE alerts (
  id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id       INT UNSIGNED NOT NULL,
  city          VARCHAR(100) NULL,
  category_id   TINYINT UNSIGNED NULL,
  min_price     DECIMAL(12,2) NULL,
  max_price     DECIMAL(12,2) NULL,
  min_area      DECIMAL(8,2)  NULL,
  max_energy    ENUM('A','B','C','D','E','F','G') NULL,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_alerts_user     FOREIGN KEY (user_id)     REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_alerts_category FOREIGN KEY (category_id) REFERENCES property_categories(id),
  INDEX idx_alerts_user (user_id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 9. INTERNAL MESSAGING (conversation client <-> agent about a property)
-- ---------------------------------------------------------------------
CREATE TABLE conversations (
  id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  property_id INT UNSIGNED NULL,            -- property context (may be NULL)
  client_id   INT UNSIGNED NOT NULL,
  agent_id    INT UNSIGNED NOT NULL,
  -- Soft-delete flags: the row is removed only once BOTH parties have deleted it.
  client_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  agent_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_conv_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE SET NULL,
  CONSTRAINT fk_conv_client   FOREIGN KEY (client_id)   REFERENCES users(id),
  CONSTRAINT fk_conv_agent    FOREIGN KEY (agent_id)    REFERENCES users(id),
  INDEX idx_conv_client (client_id),
  INDEX idx_conv_agent (agent_id)
) ENGINE=InnoDB;

CREATE TABLE messages (
  id              INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  conversation_id INT UNSIGNED NOT NULL,
  sender_id       INT UNSIGNED NOT NULL,
  body            TEXT NOT NULL,
  is_read         BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_msg_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  CONSTRAINT fk_msg_sender       FOREIGN KEY (sender_id)       REFERENCES users(id),
  INDEX idx_msg_conversation (conversation_id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 10. SCHEDULING : property VISITS + internal MEETINGS
-- ---------------------------------------------------------------------
CREATE TABLE visits (
  id           INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  property_id  INT UNSIGNED NOT NULL,
  client_id    INT UNSIGNED NOT NULL,
  agent_id     INT UNSIGNED NOT NULL,
  scheduled_at DATETIME NOT NULL,
  status       ENUM('REQUESTED','CONFIRMED','CANCELLED','COMPLETED') NOT NULL DEFAULT 'REQUESTED',
  notes        VARCHAR(500) NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_visit_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE,
  CONSTRAINT fk_visit_client   FOREIGN KEY (client_id)   REFERENCES users(id),
  CONSTRAINT fk_visit_agent    FOREIGN KEY (agent_id)    REFERENCES users(id),
  INDEX idx_visit_agent_date (agent_id, scheduled_at)
) ENGINE=InnoDB;

CREATE TABLE meetings (
  id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  agency_id     SMALLINT UNSIGNED NOT NULL,
  organizer_id  INT UNSIGNED NOT NULL,
  title         VARCHAR(150) NOT NULL,
  description   TEXT NULL,
  start_at      DATETIME NOT NULL,
  end_at        DATETIME NOT NULL,
  location      VARCHAR(150) NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_meeting_agency    FOREIGN KEY (agency_id)    REFERENCES agencies(id),
  CONSTRAINT fk_meeting_organizer FOREIGN KEY (organizer_id) REFERENCES users(id)
) ENGINE=InnoDB;

CREATE TABLE meeting_participants (
  meeting_id INT UNSIGNED NOT NULL,
  user_id    INT UNSIGNED NOT NULL,
  PRIMARY KEY (meeting_id, user_id),
  CONSTRAINT fk_mp_meeting FOREIGN KEY (meeting_id) REFERENCES meetings(id) ON DELETE CASCADE,
  CONSTRAINT fk_mp_user    FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 11. SALE FILES (transaction tracking)
-- ---------------------------------------------------------------------
CREATE TABLE sale_files (
  id              INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  property_id     INT UNSIGNED NOT NULL,
  buyer_id        INT UNSIGNED NOT NULL,
  agent_id        INT UNSIGNED NOT NULL,
  status          ENUM('OFFER','PRELIMINARY_CONTRACT','DEED','COMPLETED','CANCELLED') NOT NULL DEFAULT 'OFFER',
  negotiated_price DECIMAL(12,2) NULL,
  offer_date      DATE NULL,
  contract_date   DATE NULL,        -- "compromis"
  deed_date       DATE NULL,        -- "acte authentique"
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_sale_property FOREIGN KEY (property_id) REFERENCES properties(id),
  CONSTRAINT fk_sale_buyer    FOREIGN KEY (buyer_id)    REFERENCES users(id),
  CONSTRAINT fk_sale_agent    FOREIGN KEY (agent_id)    REFERENCES users(id),
  INDEX idx_sale_status (status)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 12. DATA / AI MODULE TABLES
-- ---------------------------------------------------------------------

-- View log -> time series for "popular properties"
-- and demand analysis per area.
CREATE TABLE property_views (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  property_id INT UNSIGNED NOT NULL,
  user_id     INT UNSIGNED NULL,            -- NULL for anonymous visitor
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_view_property FOREIGN KEY (property_id) REFERENCES properties(id) ON DELETE CASCADE,
  CONSTRAINT fk_view_user     FOREIGN KEY (user_id)     REFERENCES users(id) ON DELETE SET NULL,
  INDEX idx_view_property_date (property_id, created_at)
) ENGINE=InnoDB;

-- Price-per-m² history by city/category -> market trends.
-- Populated by the Python microservice (aggregating sales).
CREATE TABLE price_history (
  id              INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  city            VARCHAR(100) NOT NULL,
  category_id     TINYINT UNSIGNED NOT NULL,
  avg_price_per_m2 DECIMAL(10,2) NOT NULL,
  sales_count     INT UNSIGNED NOT NULL DEFAULT 0,
  period          DATE NOT NULL,             -- first day of the month
  CONSTRAINT fk_history_category FOREIGN KEY (category_id) REFERENCES property_categories(id),
  UNIQUE KEY uq_history (city, category_id, period)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- 13. SELLER APPLICATIONS
--     A buyer applies to become a seller; an agent reviews & approves.
--     Approval promotes the user's role from BUYER to SELLER.
-- ---------------------------------------------------------------------
CREATE TABLE seller_applications (
  id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     INT UNSIGNED NOT NULL,
  status      ENUM('PENDING','APPROVED','REJECTED') NOT NULL DEFAULT 'PENDING',
  motivation  VARCHAR(500) NULL,
  reviewed_by INT UNSIGNED NULL,                 -- agent who reviewed it
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  reviewed_at TIMESTAMP NULL,
  CONSTRAINT fk_sa_user     FOREIGN KEY (user_id)     REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_sa_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
  INDEX idx_sa_status (status),
  INDEX idx_sa_user (user_id)
) ENGINE=InnoDB;

-- =====================================================================
--  MINIMAL REFERENCE DATA (SEED)
-- =====================================================================
INSERT INTO departments (name, description) VALUES
  ('Management',     'Strategic steering'),
  ('Sales',          'Property buying and selling'),
  ('Comm. & Mktg',   'Communication and marketing'),
  ('Admin/HR',       'Administration and human resources'),
  ('IT & Support',   'Technical maintenance and administration');

INSERT INTO roles (code, label, is_internal) VALUES
  ('VISITOR',  'Unauthenticated visitor', FALSE),
  ('BUYER',    'Buyer client',            FALSE),
  ('SELLER',   'Seller client',           FALSE),
  ('AGENT',    'Sales agent',             TRUE),
  ('DIRECTOR', 'Agency director',         TRUE),
  ('HQ',       'Head-office staff',       TRUE),
  ('IT',       'IT and Support',          TRUE);

INSERT INTO property_categories (sector, label) VALUES
  ('RESIDENTIAL','House'),
  ('RESIDENTIAL','Apartment'),
  ('RESIDENTIAL','Studio'),
  ('RESIDENTIAL','Land'),
  ('COMMERCIAL','Office'),
  ('COMMERCIAL','Retail space'),
  ('COMMERCIAL','Warehouse');

-- Access matrix (exact translation of the brief's table)
-- Rows = requester department, value = level on target department.
-- 1=Management 2=Sales 3=Comm&Mktg 4=Admin/HR 5=IT
INSERT INTO department_permissions (requester_department_id, target_department_id, access_level) VALUES
  (1,1,'READ_WRITE'),(1,2,'READ'),(1,3,'READ'),(1,4,'READ'),(1,5,'READ'),
  (2,1,'NONE'),(2,2,'READ_WRITE'),(2,3,'READ'),(2,4,'NONE'),(2,5,'NONE'),
  (3,1,'NONE'),(3,2,'READ'),(3,3,'READ_WRITE'),(3,4,'NONE'),(3,5,'NONE'),
  (4,1,'NONE'),(4,2,'READ'),(4,3,'READ'),(4,4,'READ_WRITE'),(4,5,'NONE'),
  (5,1,'NONE'),(5,2,'READ'),(5,3,'READ'),(5,4,'NONE'),(5,5,'READ_WRITE');
