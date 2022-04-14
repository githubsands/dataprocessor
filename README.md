# data-processing

## Improvements:

- [ ]TODO: Have more refined error handling rather then using bools
- [ ]TODO: Don't use a channel and a ring bufer use either or
- [ ]TODO: If using a ring buffer possibly use atomics over mutexes in a custom ring buffer implementaion
- [ ]TODO: Use a different fd for each type of sensor in realistic production I assume we'd use network IO for new events from blockchains
- [ ]TODO: Have each sensor in its own service or a sensor per Processor not all sensors in one map
- [ ]TODO: With the hashmap within processor possibly shard the mutexes across sensors
- [ ]TODO: With the hashmap within processor possibly use the sync.hashmap
- [ ]TODO: Rather then using gonum make a custom average ring buffer that adds the floats together as they come in rather then making the float array then adding them together later
- [ ]TODO: Dispatching by string format is annoying possibly a good way to wrangle bytes here? and constant string manipulation is HEAVY
- [ ]TODO: Program needs to beable to handle a change in reference
- [ ]TODO: Program needs to beable buffer incoming logs when a reference change is occuring
- [ ]TODO: Swap file IO for networkIO for incoming logs
- [ ]TODO: Apply structure packing to batch when cache data structure is decided upon
