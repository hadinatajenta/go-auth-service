CREATE TABLE "users" (
  "id" BIGSERIAL PRIMARY KEY,
  "first_name" varchar(50),
  "last_name" varchar(50),
  "username" varchar(50) UNIQUE,
  "email" varchar(100) UNIQUE,
  "password" varchar(255),
  "last_login" timestamp,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "user_sessions" (
  "id" BIGSERIAL PRIMARY KEY,
  "user_id" bigint,
  "token" varchar(255),
  "ip_address" varchar(45),
  "user_agent" varchar(255),
  "expired_at" timestamp,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "roles" (
  "id" BIGSERIAL PRIMARY KEY,
  "name" varchar(50) UNIQUE,
  "description" varchar(255),
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "user_roles" (
  "id" BIGSERIAL PRIMARY KEY,
  "user_id" bigint,
  "role_id" bigint,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "permissions" (
  "id" BIGSERIAL PRIMARY KEY,
  "name" varchar(100) UNIQUE,
  "description" varchar(255),
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "role_permissions" (
  "id" BIGSERIAL PRIMARY KEY,
  "role_id" bigint,
  "permission_id" bigint,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "menus" (
  "id" BIGSERIAL PRIMARY KEY,
  "name" varchar(100),
  "description" varchar(255),
  "path" varchar(255),
  "icon" varchar(50),
  "parent_id" bigint,
  "sort_order" int,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "menu_permissions" (
  "id" BIGSERIAL PRIMARY KEY,
  "menu_id" bigint,
  "permission_id" bigint,
  "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE "user_sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "user_roles" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "user_roles" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");
ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");
ALTER TABLE "role_permissions" ADD FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id");
ALTER TABLE "menu_permissions" ADD FOREIGN KEY ("menu_id") REFERENCES "menus" ("id");
ALTER TABLE "menu_permissions" ADD FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id");

CREATE UNIQUE INDEX ON "user_roles" ("user_id", "role_id");
CREATE UNIQUE INDEX ON "role_permissions" ("role_id", "permission_id");
CREATE UNIQUE INDEX ON "menu_permissions" ("menu_id", "permission_id");
