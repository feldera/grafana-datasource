# This script uses the Feldera API to create and start a pipeline using SQL
# program in 'test.sql'.

from feldera import FelderaClient, PipelineBuilder

sql = open("grafana.sql").read()

client = FelderaClient('http://localhost:8080')

print('Starting pipeline')
pipeline = PipelineBuilder(client, 'grafana', sql).create_or_replace()
pipeline.start()
