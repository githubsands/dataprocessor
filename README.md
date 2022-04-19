# data-processing

data-processing processes data in either batches or streams depending on the sensor noted in a received log event
currently the program can be ran with terminal input or simulation input

to run with simulation input: `data-processing -simulation true -sensors 5 -samples 10`

this creates a total of 10 sensors: 5 temperature, 5 humidity and 250(5*5*10) samples. it then sends them to the processor
through a io pipe and outputs the logs to `/log/*`

currently the program can handle 100,000 samples per batch with a total of 5 batches
after this limit the completion of batches is unknown/exceeds 5 minutes (`./data-processing -simulation true -sensors 5 -samples 100000`)

additionaly the program can handle stream processing with the above limit listed

future changes are to address swapping local input with network input as noted in with the `io.go` interface and files within the `io/grpc` and more included in improvements section below

currently the distribution curve that generates temp and humidity events needs to be tweaked to see all possible outputs (discard, very precise, etc..)

## Improvements:

- [ ]TODO: Don't use a channel and a ring buffer use either or
- [ ]TODO: If using a ring buffer possibly use atomics over mutexes in a custom ring buffer implementaion
- [ ]TODO: Have each sensor in its own service
- [ ]TODO: With the hashmap within processor possibly shard the mutexes across sensors
- [ ]TODO: With the hashmap within processor possibly use the sync.hashmap
- [ ]TODO: Rather then using gonum make a custom average ring buffer that adds the floats together as they come in rather then making the float array then adding them together later
- [ ]TODO: Program needs to beable to handle a change in reference
- [ ]TODO: Swap fileIO with networkIO for incoming logs
- [ ]TODO: Streamers need to reject all future logs if its log has already been processed
- [ ]Fix main.go waitgroups

....and more
