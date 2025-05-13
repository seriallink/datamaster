/*
Created: 5/6/2025
Modified: 5/13/2025
Project: Data Master
Model: DataMaster
Author: Marcelo Monaco
Database: Aurora PostgreSQL
*/

-- Create schemas section -------------------------------------------------

CREATE SCHEMA dm
;

-- Create user data types section -------------------------------------------------

CREATE TYPE dm.delivery_status_enum AS ENUM
  ( 'pending', 'in_transit', 'delivered', 'failed', 'canceled' )
;

CREATE TYPE dm.gender_enum AS ENUM
  ( 'male', 'female', 'undisclosed' )
;

CREATE TYPE dm.purchase_status_enum AS ENUM
  ( 'pending', 'confirmed', 'shipped', 'delivered', 'canceled' )
;

CREATE TYPE dm.sentiment_enum AS ENUM
  ( 'positive', 'neutral', 'negative' )
;

CREATE TYPE dm.tracking_status_enum AS ENUM
  (   'created', 'picked_up', 'in_transit', 'delivered', 'delayed', 'failed_delivery', 'returned','lost', 'damaged' )
;

-- Create tables section -------------------------------------------------

-- Table dm.category

CREATE TABLE dm.category
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

ALTER TABLE dm.category ADD CONSTRAINT PK_category PRIMARY KEY (category_id)
;

-- Table dm.courier

CREATE TABLE dm.courier
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

ALTER TABLE dm.courier ADD CONSTRAINT pk_courier PRIMARY KEY (courier_id)
;

-- Table dm.customer

CREATE TABLE dm.customer
(
  customer_id UUID NOT NULL,
  customer_name Text NOT NULL,
  customer_email Text NOT NULL,
  customer_phone Text NOT NULL,
  gender dm.gender_enum NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm.customer ADD CONSTRAINT pk_customer PRIMARY KEY (customer_id)
;

-- Table dm.delivery

CREATE TABLE dm.delivery
(
  delivery_id UUID NOT NULL,
  purchase_id UUID NOT NULL,
  courier_id UUID NOT NULL,
  delivery_status dm.delivery_status_enum NOT NULL,
  delivery_estimated_date Timestamptz NOT NULL,
  delivery_date Timestamptz,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_delivery_order_id ON dm.delivery (purchase_id)
;

CREATE INDEX ix_delivery_courier_id ON dm.delivery (courier_id)
;

ALTER TABLE dm.delivery ADD CONSTRAINT PK_delivery PRIMARY KEY (delivery_id)
;

-- Table dm.product

CREATE TABLE dm.product
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

CREATE INDEX ix_product_category_id ON dm.product (category_id)
;

ALTER TABLE dm.product ADD CONSTRAINT pk_product PRIMARY KEY (product_id)
;

-- Table dm.purchase

CREATE TABLE dm.purchase
(
  purchase_id UUID NOT NULL,
  customer_id UUID NOT NULL,
  purchase_status dm.purchase_status_enum NOT NULL,
  total_value Numeric NOT NULL,
  purchase_date Timestamptz NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_purchase_customer_id ON dm.purchase (customer_id)
;

ALTER TABLE dm.purchase ADD CONSTRAINT pk_purchase PRIMARY KEY (purchase_id)
;

-- Table dm.purchase_item

CREATE TABLE dm.purchase_item
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

ALTER TABLE dm.purchase_item ADD CONSTRAINT pk_purchase_item PRIMARY KEY (purchase_id,product_id)
;

-- Table dm.review

CREATE TABLE dm.review
(
  review_id UUID NOT NULL,
  product_id UUID NOT NULL,
  customer_id UUID NOT NULL,
  review_text Text NOT NULL,
  rating Smallint NOT NULL
    CONSTRAINT ck_product_review_rating CHECK (rating BETWEEN 1 AND 5),
  review_date Timestamptz NOT NULL,
  sentiment_label dm.sentiment_enum,
  sentiment_score Numeric(3,2),
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_review_product_id ON dm.review (product_id)
;

CREATE INDEX ix_review_customer_id ON dm.review (customer_id)
;

ALTER TABLE dm.review ADD CONSTRAINT pk_review PRIMARY KEY (review_id)
;

-- Table dm.tracking

CREATE TABLE dm.tracking
(
  tracking_id UUID NOT NULL,
  delivery_id UUID NOT NULL,
  notes Text,
  tracking_status dm.tracking_status_enum NOT NULL,
  event_timestamp Timestamptz NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_tracking_delivery_id ON dm.tracking (delivery_id)
;

ALTER TABLE dm.tracking ADD CONSTRAINT pk_tracking PRIMARY KEY (tracking_id)
;

-- Create foreign keys (relationships) section -------------------------------------------------

ALTER TABLE dm.purchase
  ADD CONSTRAINT fk_purchase_customer
    FOREIGN KEY (customer_id)
    REFERENCES dm.customer (customer_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.delivery
  ADD CONSTRAINT fk_delivery_purchase
    FOREIGN KEY (purchase_id)
    REFERENCES dm.purchase (purchase_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.product
  ADD CONSTRAINT fk_product_category
    FOREIGN KEY (category_id)
    REFERENCES dm.category (category_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.review
  ADD CONSTRAINT fk_review_product
    FOREIGN KEY (product_id)
    REFERENCES dm.product (product_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.review
  ADD CONSTRAINT fk_review_customer
    FOREIGN KEY (customer_id)
    REFERENCES dm.customer (customer_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.purchase_item
  ADD CONSTRAINT fk_purchase_item_purchase
    FOREIGN KEY (purchase_id)
    REFERENCES dm.purchase (purchase_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.purchase_item
  ADD CONSTRAINT fk_purchase_item_product
    FOREIGN KEY (product_id)
    REFERENCES dm.product (product_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.delivery
  ADD CONSTRAINT fk_delivery_courier
    FOREIGN KEY (courier_id)
    REFERENCES dm.courier (courier_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm.tracking
  ADD CONSTRAINT fk_tracking_delivery
    FOREIGN KEY (delivery_id)
    REFERENCES dm.delivery (delivery_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;
