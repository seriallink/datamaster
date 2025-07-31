import boto3
import json
import sys
import time

from awsglue.utils import getResolvedOptions
from pyspark.sql import SparkSession
from pyspark.sql.types import *
from pyspark.sql.functions import col

def log(msg):
    print(f"[BENCHMARK] {msg}", flush=True)

def main():
    start_time = time.time()

    log("Starting SparkSession...")
    spark = SparkSession.builder \
        .appName("GlueBenchmark") \
        .config("spark.hadoop.mapreduce.fileoutputcommitter.algorithm.version", "1") \
        .config("spark.speculation", "false") \
        .getOrCreate()

    log("Reading environment variables...")
    args = getResolvedOptions(sys.argv, [
        'BENCHMARK_BUCKET',
        'BENCHMARK_INPUT',
        'BENCHMARK_OUTPUT',
        'BENCHMARK_RESULT'
    ])

    bucket = args['BENCHMARK_BUCKET']
    input_key = args['BENCHMARK_INPUT']
    output_key = args['BENCHMARK_OUTPUT']
    result_key = args['BENCHMARK_RESULT']

    s3_path_in = f"s3://{bucket}/{input_key}"
    s3_path_out = f"s3://{bucket}/{output_key}"

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

    log("Evaluating DataFrame columns...")
    df = df.select([col(c) for c in df.columns])

    log(f"Writing Parquet to {s3_path_out}")
    parquet_start = time.time()
    df.write.mode("overwrite").parquet(s3_path_out, compression="snappy")
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
