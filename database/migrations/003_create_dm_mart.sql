/*
Created: 5/6/2025
Modified: 6/6/2025
Project: Data Master
Model: DataMaster
Author: Marcelo Monaco
Database: Aurora PostgreSQL
*/

-- Create schemas section -------------------------------------------------

CREATE SCHEMA dm_mart
;
COMMENT ON SCHEMA dm_mart IS 'Analytical data models and summary tables optimized for reporting and dashboarding. Used to populate the Gold layer in the data lake.'
;

-- Create tables section -------------------------------------------------

-- Table dm_mart.top_beers_by_rating

CREATE TABLE dm_mart.top_beers_by_rating
(
  beer_id Bigint NOT NULL,
  beer_name Text NOT NULL,
  beer_style Text NOT NULL,
  avg_rating Numeric NOT NULL,
  total_reviews Bigint NOT NULL,
  review_year Smallint NOT NULL,
  review_month Smallint NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

-- Table dm_mart.top_breweries_by_rating

CREATE TABLE dm_mart.top_breweries_by_rating
(
  brewery_id Bigint NOT NULL,
  brewery_name Text NOT NULL,
  avg_rating Numeric NOT NULL,
  total_reviews Bigint NOT NULL,
  review_year Smallint NOT NULL,
  review_month Smallint NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

-- Table dm_mart.state_by_review_volume

CREATE TABLE dm_mart.state_by_review_volume
(
  state Text NOT NULL,
  total_reviews Bigint NOT NULL,
  review_year Smallint NOT NULL,
  review_month Smallint NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

-- Table dm_mart.top_styles_by_popularity

CREATE TABLE dm_mart.top_styles_by_popularity
(
  beer_style Text NOT NULL,
  total_reviews Bigint NOT NULL,
  avg_rating Numeric NOT NULL,
  review_year Smallint NOT NULL,
  review_month Smallint NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

-- Table dm_mart.top_drinkers

CREATE TABLE dm_mart.top_drinkers
(
  profile_id Text NOT NULL,
  state Text NOT NULL,
  total_reviews Bigint NOT NULL,
  avg_rating Numeric NOT NULL,
  review_year Smallint NOT NULL,
  review_month Smallint NOT NULL
)
WITH (
  autovacuum_enabled=true)
;

