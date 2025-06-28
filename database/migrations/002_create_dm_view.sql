/*
Created: 5/6/2025
Modified: 6/27/2025
Project: Data Master
Model: DataMaster
Author: Marcelo Monaco
Database: Aurora PostgreSQL 9.5
*/

-- Create schemas section -------------------------------------------------

CREATE SCHEMA dm_view
;
COMMENT ON SCHEMA dm_view IS 'Normalized and enriched views derived from core tables. Represents cleaned and deduplicated data used to populate the Silver layer in the data lake.'
;

-- Create tables section -------------------------------------------------

-- Table dm_view.brewery

CREATE TABLE dm_view.brewery
(
  brewery_id Bigint NOT NULL,
  brewery_name Text NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL,
  deleted_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_view.brewery ADD CONSTRAINT pk_brewery PRIMARY KEY (brewery_id)
;

-- Table dm_view.beer

CREATE TABLE dm_view.beer
(
  beer_id Bigint NOT NULL,
  beer_name Text NOT NULL,
  beer_style Text NOT NULL,
  beer_abv Numeric,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL,
  deleted_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_view.beer ADD CONSTRAINT pk_beer PRIMARY KEY (beer_id)
;

-- Table dm_view.profile

CREATE TABLE dm_view.profile
(
  profile_id Bigint NOT NULL,
  profile_name Text NOT NULL,
  email Text NOT NULL,
  state Text NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL,
  deleted_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_view.profile ADD CONSTRAINT pk_profile PRIMARY KEY (profile_id)
;

-- Table dm_view.review

CREATE TABLE dm_view.review
(
  review_id Bigint NOT NULL,
  brewery_id Bigint NOT NULL,
  beer_id Bigint NOT NULL,
  profile_id Bigint NOT NULL,
  review_overall Numeric,
  review_aroma Numeric,
  review_appearance Numeric,
  review_palate Numeric,
  review_taste Numeric,
  review_time Timestamptz NOT NULL,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL,
  deleted_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_view.review ADD CONSTRAINT PK_review PRIMARY KEY (review_id)
;

-- Table dm_view.review_flat

CREATE TABLE dm_view.review_flat
(
  review_id Bigint NOT NULL,
  brewery_id Bigint NOT NULL,
  beer_id Bigint NOT NULL,
  profile_id Bigint NOT NULL,
  brewery_name Text,
  beer_name Text,
  beer_style Text,
  beer_abv Numeric,
  profile_name Text,
  email Text,
  state Text,
  review_overall Numeric,
  review_aroma Numeric,
  review_appearance Numeric,
  review_palate Numeric,
  review_taste Numeric,
  review_time Timestamptz,
  created_at Timestamptz NOT NULL,
  updated_at Timestamptz NOT NULL,
  deleted_at Timestamptz NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_view.review_flat ADD CONSTRAINT pk_review_flat PRIMARY KEY (review_id)
;

