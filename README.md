[Documentation](https://godoc.org/github.com/hermanschaaf/cedict)

CEDict Parser in Go
-------------------

Package cedict provides a parser / tokenizer for reading entries from the [CC-CEDict
Chinese dictionary project](http://www.mdbg.net/chindict/chindict.php?page=cedict).

### Installation

Assuming you have Go installed, installation is as easy as running:

```
go get github.com/hermanschaaf/cedict
```

You will need a copy of the CEDict dictionary text file. You can [download CEDict from MDBG.net](http://www.mdbg.net/chindict/chindict.php?page=cedict). Extract the file somewhere you want to use it from, and then follow the usage instructions below.

### Usage

Tokenizing is done by creating a `CEDict` for an `io.Reader` `r`. It is the
caller's responsibility to ensure that `r` provides a CEDict-formatted dictionary.

        import "github.com/hermanschaaf/cedict"

        ...

        c := cedict.New(r) // r is an io.Reader to the cedict file

Given a `CEDict` called `c`, the dictionary is tokenized by repeatedly calling `c.NextEntry()`,
which parses until it reaches the next entry, or an error if no more entries are found:

```
    for {
        err := c.NextEntry()
        if err != nil {
            break
        }
        entry := c.Entry()
        fmt.Println(entry.Simplified, entry.Definitions[0])
    }
```

To retrieve the current entry, the `Entry` method can be called. There is also
a lower-level API available, using the `bufio.Scanner` `Scan` method. Using this
lower-level API is the recommended way to read comments from the `CEDict`, should
that be necessary.

### Documentation

Full documentation can be found at [https://godoc.org/github.com/hermanschaaf/japanese](https://godoc.org/github.com/hermanschaaf/cedict)