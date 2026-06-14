-- =====================================================================
--  MINIMAL REFERENCE DATA
--  Essential data for the application to function (roles, categories, matrix)
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
