/*
Created: 5/6/2025
Modified: 6/10/2025
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

-- Create tables section -------------------------------------------------

-- Table dm_core.brewery

CREATE TABLE dm_core.brewery
(
  brewery_id Bigint NOT NULL,
  brewery_name Text NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.brewery ADD CONSTRAINT pk_brewery PRIMARY KEY (brewery_id)
;

-- Table dm_core.beer

CREATE TABLE dm_core.beer
(
  beer_id Bigint NOT NULL,
  brewery_id Bigint NOT NULL,
  beer_name Text NOT NULL,
  beer_style Text NOT NULL,
  beer_abv Numeric
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_beer_brewery ON dm_core.beer (brewery_id)
;

ALTER TABLE dm_core.beer ADD CONSTRAINT pk_beer PRIMARY KEY (beer_id)
;

-- Table dm_core.profile

CREATE TABLE dm_core.profile
(
  profile_id Bigint NOT NULL,
  profile_name Text NOT NULL,
  email Text NOT NULL,
  state Text NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

ALTER TABLE dm_core.profile ADD CONSTRAINT pk_profile PRIMARY KEY (profile_id)
;

-- Table dm_core.review

CREATE TABLE dm_core.review
(
  review_id Bigint NOT NULL,
  beer_id Bigint NOT NULL,
  profile_id Bigint NOT NULL,
  review_overall Numeric,
  review_aroma Numeric,
  review_appearance Numeric,
  review_palate Numeric,
  review_taste Numeric,
  review_time Bigint
)
WITH (
  autovacuum_enabled=true)
;

CREATE INDEX ix_review_beer ON dm_core.review (beer_id)
;

CREATE INDEX ix_review_profile ON dm_core.review (profile_id)
;

ALTER TABLE dm_core.review ADD CONSTRAINT PK_review PRIMARY KEY (review_id)
;

-- Create foreign keys (relationships) section -------------------------------------------------

ALTER TABLE dm_core.beer
  ADD CONSTRAINT fk_beer_brewery
    FOREIGN KEY (brewery_id)
    REFERENCES dm_core.brewery (brewery_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.review
  ADD CONSTRAINT fk_review_beer
    FOREIGN KEY (beer_id)
    REFERENCES dm_core.beer (beer_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

ALTER TABLE dm_core.review
  ADD CONSTRAINT fk_review_profile
    FOREIGN KEY (profile_id)
    REFERENCES dm_core.profile (profile_id)
      ON DELETE NO ACTION
      ON UPDATE NO ACTION
;

