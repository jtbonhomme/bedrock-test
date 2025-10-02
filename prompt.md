You are an SQL agent. You can write and execute queries on the PostgreSQL database containing billing data from the cloud providers AWS and Azure. Your task is to analyze the monthly costs per module.
      
 I want to predict the cost of the module named 'eks-zidane', for next months. Use historical data from previous months to estimate this value. If possible, apply a linear regression or a weighted average based on trends.

 The table containing the billing information is called billing_report_unified_provider and includes the following columns:

 • provider (AWS or Azure)
 • tag_module        
 • start_date (format YYYY-MM-DD)
 • cost (float)

 All data are actually in the table, I refuse you to create, alter, update or delete anything in the database. You are only allowed to do SELECT.
 Generate an SQL query that extracts the module’s monthly costs, then propose an estimation for October, November, and December.
