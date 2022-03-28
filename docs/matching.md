# Matching

## What happens if you remove the go-command from the Seek call in the main function?

### Hypothesis
The program will reach a deadlock since the first `Seek()` call will send on the match channel and wait for it's value to be read. This however will never happen. Since it's the only thread runnig.

### Result
No nevermind. It works just fine. I forgot that the channel is buffered. This means that the thread doesn't have to wait for someone to recieve before it can move on. The for loop can then continue and the next call to `Seek()` will read the value.


## What happens if you switch the declaration wg := new(sync.WaitGroup) to var wg sync.WaitGroup and the parameter wg *sync.WaitGroup to wg sync.WaitGroup?

### Hypothesis
This time the program will actually reach a deadlock. This is due to how pass by value works in Go. `wg := new(sync.WaitGroup)` allocates a "zero valued" WaitGroup and returns the pointer to it, meaning wg is a pointer to a WaitGroup. When declaring wg as `var wg sync.WaitGroup` wg instead is the WaitGroup as a value. When a function takes a value in Go it is copied to another part of memory and any changes to it will therefore not be applied to the original value. In this case it means that the `wg.Done()` won't affect the original WaitGroup. But this means that the program will wait indefinetely when reaching `wg.Wait()`.

### Result
Just as I predicted the program reached a deadlock.  AT the `wg.Wait()`.
    
## What happens if you remove the buffer on the channel match?

### Hypothesis
Once again I think the program should reach a deadlock. Removing the buffer will make `Seek()` functions that send on the match channel have to wait before someone other thread reads their value. This works fine until the last  `Seek()` (if odd number of people). This last `Seek()` will then wait for its value to be read before calling `wg.Done()`. The issue is that the main routine waits at `wg.Wait()` before reading the value on the match channel. So main routine waits for the go routine to say it's done and the go routine waits for the main routine to read its value.

### Result
Like I said.

## What happens if you remove the default-case from the case-statement in the main function?

### Hypothesis
The program won't exit if there's an even ammount of people. Then there will be nothing to read on the match channel and the select statement will wait indefinetely for a value on the match channel. In other words a deadlock.

### Result
Once again the hypothesis was correct