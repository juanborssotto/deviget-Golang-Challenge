# Golang-Challenge

## Steps
* Read cache.go and cache_test.go to comprehend the problem.
* First TODO: noticed that a way to know when a price was cached was necessary so proceed
  to create a custom price struct. Updated function to save that time along the price itself and then
  implemented the if statement. Did a manual test in the main function to check everything was fine.
* Second TODO: noticed that GetPriceFor was reading and writing a shared resource, so proceed to add
  a mutex to lock before those operations. Then added the parallelization logic to GetPricesFor and same 
  as before, tested it in the main function.
