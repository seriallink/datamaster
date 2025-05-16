/*
Created: 5/6/2025
Modified: 5/15/2025
Project: Data Master
Model: DataMaster
Author: Marcelo Monaco
Database: Aurora PostgreSQL 9.5
*/

-- Create schemas section -------------------------------------------------

CREATE SCHEMA dm_core
;
COMMENT ON SCHEMA dm_core IS 'Raw transactional data schema. Mirrors the source systems without transformation. Used to populate the Bronze layer in the data lake.'
;

-- Create user data types section -------------------------------------------------

CREATE TYPE dm_core.delivery_status_enum AS ENUM
  ( 'pending', 'in_transit', 'delivered', 'failed', 'canceled' )
;

CREATE TYPE dm_core.gender_enum AS ENUM
  ( 'male', 'female', 'undisclosed' )
;

CREATE TYPE dm_core.purchase_status_enum AS ENUM
  ( 'pending', 'confirmed', 'shipped', 'delivered', 'canceled' )
;

CREATE TYPE dm_core.sentiment_enum AS ENUM
  ( 'positive', 'neutral', 'negative' )
;

CREATE TYPE dm_core.tracking_status_enum AS ENUM
  (   'created', 'picked_up', 'in_transit', 'delivered', 'delayed', 'failed_delivery', 'returned','lost', 'damaged' )
;

-- Create tables section -------------------------------------------------

-- Table dm_core.category

CREATE TABLE dm_core.category
(
  category_id UUID NOT NULL,
  category_name Text NOT NULL,
  category_description Text NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.category ADD CONSTRAINT PK_category PRIMARY KEY (category_id)
;

-- Table dm_core.courier

CREATE TABLE dm_core.courier
(
  courier_id UUID NOT NULL,
  courier_name Text NOT NULL,
  contact_name Text NOT NULL,
  contact_email Text NOT NULL,
  contact_phone Text NOT NULL,
  cost_per_km Numeric NOT NULL,
  average_delivery_time Interval  NOT NULL,
  reliability_score Smallint NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.courier ADD CONSTRAINT pk_courier PRIMARY KEY (courier_id)
;

-- Table dm_core.customer

CREATE TABLE dm_core.customer
(
  customer_id UUID NOT NULL,
  customer_name Text NOT NULL,
  customer_email Text NOT NULL,
  customer_phone Text NOT NULL,
  gender dm_core.gender_enum NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.customer ADD CONSTRAINT pk_customer PRIMARY KEY (customer_id)
;

-- Table dm_core.delivery

CREATE TABLE dm_core.delivery
(
  delivery_id UUID NOT NULL,
  purchase_id UUID NOT NULL,
  courier_id UUID NOT NULL,
  delivery_status dm_core.delivery_status_enum NOT NULL,
  delivery_estimated_date Timestamptz NOT NULL,
  delivery_date Timestamptz,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_delivery_order_id ON dm_core.delivery (purchase_id)
;

CREATE INDEX ix_delivery_courier_id ON dm_core.delivery (courier_id)
;

ALTER TABLE dm_core.delivery ADD CONSTRAINT PK_delivery PRIMARY KEY (delivery_id)
;

-- Table dm_core.product

CREATE TABLE dm_core.product
(
  product_id UUID NOT NULL,
  category_id UUID,
  product_name Text NOT NULL,
  current_stock Numeric NOT NULL,
  unit_price Numeric NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_product_category_id ON dm_core.product (category_id)
;

ALTER TABLE dm_core.product ADD CONSTRAINT pk_product PRIMARY KEY (product_id)
;

-- Table dm_core.purchase

CREATE TABLE dm_core.purchase
(
  purchase_id UUID NOT NULL,
  customer_id UUID NOT NULL,
  purchase_status dm_core.purchase_status_enum NOT NULL,
  total_value Numeric NOT NULL,
  purchase_date Timestamptz NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_purchase_customer_id ON dm_core.purchase (customer_id)
;

ALTER TABLE dm_core.purchase ADD CONSTRAINT pk_purchase PRIMARY KEY (purchase_id)
;

-- Table dm_core.purchase_item

CREATE TABLE dm_core.purchase_item
(
  purchase_id UUID NOT NULL,
  product_id UUID NOT NULL,
  quantity Numeric NOT NULL,
  unit_price Numeric NOT NULL,
  total_price Numeric NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.purchase_item ADD CONSTRAINT pk_purchase_item PRIMARY KEY (purchase_id,product_id)
;

-- Table dm_core.review

CREATE TABLE dm_core.review
(
  review_id UUID NOT NULL,
  product_id UUID NOT NULL,
  customer_id UUID NOT NULL,
  review_text Text NOT NULL,
  rating Smallint NOT NULL
    CONSTRAINT ck_product_review_rating CHECK (rating BETWEEN 1 AND 5),
  review_date Timestamptz NOT NULL,
  sentiment_label dm_core.sentiment_enum,
  sentiment_score Numeric(3,2),
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_review_product_id ON dm_core.review (product_id)
;

CREATE INDEX ix_review_customer_id ON dm_core.review (customer_id)
;

ALTER TABLE dm_core.review ADD CONSTRAINT pk_review PRIMARY KEY (review_id)
;

-- Table dm_core.tracking

CREATE TABLE dm_core.tracking
(
  tracking_id UUID NOT NULL,
  delivery_id UUID NOT NULL,
  notes Text,
  tracking_status dm_core.tracking_status_enum NOT NULL,
  event_timestamp Timestamptz NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_tracking_delivery_id ON dm_core.tracking (delivery_id)
;

ALTER TABLE dm_core.tracking ADD CONSTRAINT pk_tracking PRIMARY KEY (tracking_id)
;

-- Create foreign keys (relationships) section -------------------------------------------------

ALTER TABLE dm_core.purchase
  ADD CONSTRAINT fk_purchase_customer
    FOREIGN KEY (customer_id)
    REFERENCES dm_core.customer (customer_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.delivery
  ADD CONSTRAINT fk_delivery_purchase
    FOREIGN KEY (purchase_id)
    REFERENCES dm_core.purchase (purchase_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.product
  ADD CONSTRAINT fk_product_category
    FOREIGN KEY (category_id)
    REFERENCES dm_core.category (category_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.review
  ADD CONSTRAINT fk_review_product
    FOREIGN KEY (product_id)
    REFERENCES dm_core.product (product_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.review
  ADD CONSTRAINT fk_review_customer
    FOREIGN KEY (customer_id)
    REFERENCES dm_core.customer (customer_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.purchase_item
  ADD CONSTRAINT fk_purchase_item_purchase
    FOREIGN KEY (purchase_id)
    REFERENCES dm_core.purchase (purchase_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.purchase_item
  ADD CONSTRAINT fk_purchase_item_product
    FOREIGN KEY (product_id)
    REFERENCES dm_core.product (product_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.delivery
  ADD CONSTRAINT fk_delivery_courier
    FOREIGN KEY (courier_id)
    REFERENCES dm_core.courier (courier_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.tracking
  ADD CONSTRAINT fk_tracking_delivery
    FOREIGN KEY (delivery_id)
    REFERENCES dm_core.delivery (delivery_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

