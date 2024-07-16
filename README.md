# sitemap
sitemap is a small project that helps visualize website layouts.

The program scans a specified website and follows all links under the same domain using a DFS. 

After organizing the results, the program outputs a dot file which can be modified and then compiled.

the default url is set to my univeristy, which gives a fairly complex result.
```./sitemap -url="https://www.avemaria.edu"```

There are many things that need to be fixed, including a more verbose GET request and accounting for things like subdomains.

You will need to install dot in order to build the output.

```dot -Tpdf sitemap.dot -o output.pdf```

I enjoyed learning about the go programming languages and practicing some more complex algorithim and data structure concepts.

