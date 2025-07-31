import argparse
import boto3
import json
import time

from pyspark.sql import SparkSession
from pyspark.sql.types import *

def log(msg):
    print(f"[BENCHMARK] {msg}", flush=True)

def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--benchmark_bucket", required=True)
    parser.add_argument("--benchmark_input", required=True)
    parser.add_argument("--benchmark_output", required=True)
    parser.add_argument("--benchmark_result", required=True)
    return parser.parse_args()

def main():
    start_time = time.time()

    log("Reading parameters...")
    args = parse_args()
    bucket = args.benchmark_bucket
    input_key = args.benchmark_input
    output_key = args.benchmark_output
    result_key = args.benchmark_result

    s3_path_in = f"s3://{bucket}/{input_key}"
    s3_path_out = f"s3://{bucket}/{output_key}"

    log("Starting SparkSession...")
    spark = SparkSession.builder \
        .appName("dm-benchmark-emr") \
        .config("spark.sql.parquet.compression.codec", "snappy") \
        .getOrCreate()

    schema = StructType([
        StructField("review_id", LongType()),
        StructField("brewery_id", LongType()),
        StructField("beer_id", LongType()),
        StructField("profile_id", LongType()),
        StructField("review_overall", DoubleType()),
        StructField("review_aroma", DoubleType()),
        StructField("review_appearance", DoubleType()),
        StructField("review_palate", DoubleType()),
        StructField("review_taste", DoubleType()),
        StructField("review_time", LongType()),
    ])

    log(f"Reading CSV from {s3_path_in}")
    csv_start = time.time()

    df = spark.read \
        .option("header", "true") \
        .option("compression", "gzip") \
        .schema(schema) \
        .csv(s3_path_in)

    csv_end = time.time()
    csv_elapsed = csv_end - csv_start
    log(f"CSV read done in {csv_elapsed:.6f}s")

    log(f"Writing Parquet to {s3_path_out}")
    parquet_start = time.time()

    df.write \
        .mode("overwrite") \
        .parquet(s3_path_out)

    parquet_end = time.time()
    parquet_elapsed = parquet_end - parquet_start
    log(f"Parquet write done in {parquet_elapsed:.6f}s")

    total_elapsed = time.time() - start_time
    log(f"Total task time: {total_elapsed:.6f}s")

    result = {
        "start_task_time": time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime(start_time)),
        "end_task_time": time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()),
        "input_file": input_key,
        "output_file": output_key,
        "csv_read_time": int(csv_elapsed * 1e9),
        "parquet_write_time": int(parquet_elapsed * 1e9),
        "total_task_time": int(total_elapsed * 1e9),
    }

    log(f"Writing benchmark result to s3://{bucket}/{result_key}")
    boto3.client("s3").put_object(
        Bucket=bucket,
        Key=result_key,
        Body=json.dumps(result, indent=2).encode("utf-8")
    )

    log("Stopping SparkSession.")
    spark.stop()

    log("Benchmark completed.")

if __name__ == "__main__":
    main()
