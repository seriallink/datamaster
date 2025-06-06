/*
Created: 5/6/2025
Modified: 6/6/2025
Project: Data Master
Model: DataMaster
Author: Marcelo Monaco
Database: Aurora PostgreSQL
*/

-- Create schemas section -------------------------------------------------

CREATE SCHEMA dm_view
;
COMMENT ON SCHEMA dm_view IS 'Normalized and enriched views derived from core tables. Represents cleaned and deduplicated data used to populate the Silver layer in the data lake.'
;

-- Create tables section -------------------------------------------------

-- Table dm_view.review_normalized

CREATE TABLE dm_view.review_normalized
(
  review_id Bigint NOT NULL,
  brewery_id Bigint NOT NULL,
  beer_id Bigint NOT NULL,
  profile_id Text NOT NULL,
  brewery_name Text,
  beer_name Text,
  beer_style Text,
  beer_abv Numeric,
  email Text,
  state Text,
  review_overall Numeric,
  review_aroma Numeric,
  review_appearance Numeric,
  review_palate Numeric,
  review_taste Numeric,
  review_time Timestamptz
)
WITH (
  autovacuum_enabled=true)
;

