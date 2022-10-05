'use strict'

import {List, Item} from 'linked-list'
import {randomBytes} from 'crypto' ;
import {parentPort} from 'worker_threads'

let memoryHeld = new List()

parentPort.on("message",(payload, node) => {
    console.log(`Mem: new event ${JSON.stringify(payload)}`)
    if (payload.action && payload.action === "chew") {
        memoryHeld = chewMemory(payload.totalMemoryUsage)
    } else {
        memoryHeld = new List()
        gc()
        console.log(`Mem: got a ${payload.action} freeing memory`)
        process.exit()
    }
})

class MemBlock extends Item {
    constructor(value) {
        super()
        this.value = value
    }

    toString() {
        return this.value
    }
}

function chewMemory(totalMemoryUsage) {
    console.log(`mem: before allocation memory used: ${JSON.stringify(process.memoryUsage())}`)
    console.log(`mem: trying to allocate ${totalMemoryUsage} M`)
    const mbBytes = 1000 * 1024
    const bufferSize = 1024 * 4

    let memoryList = new List()
    try {
        while ((process.memoryUsage().rss / mbBytes) < totalMemoryUsage) {
            memoryList.append(new MemBlock(new MemBlock(randomBytes(bufferSize).toString('hex'))))
        }
    } catch (e) {
        console.log(`Exception caught while allocating mem ... stopping:${e.toString()}`)
    }
    console.log(`mem: memory used array length was ${memoryList.size}  amount of memory used ${JSON.stringify(process.memoryUsage())}`)
    return memoryList
}

function gc() {
    try {
        if (global.gc) {
            global.gc()
        } else {
            console.log("There is no gc exposed to the worker thread.")
        }
    } catch (e) {
        console.log("Tried to garbage collect Failed please start with the option '--expose-gc'")
    }
}
