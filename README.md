# dvmapp

Server for Project Die Virtuelle Mittagsfrau.

## Contents

### Homepage

On startup, make an inventory of media files (media/dvm-..) for images and
videos. At template execution time, choose a random element.

Random media and link to random read and write page (/r/080912 or /w/121708 ...).


* an random animation
* link to the details [write] page for the current animation
* some stats and random sentences from stories
* number of translations, invitation to translate


The animations are numbered:

```
mittagsfrau.de/i/011421.gifv
mittagsfrau.de/i/211413.jpg
```

```
mittagsfrau.de
```

### Details

Various details.

```
mittagsfrau.de/w/345 [form, submit text]
mittagsfrau.de/r/345
mittagsfrau.de/t/345 [form, submit translation]
```

Link to a certain story:

```
mittagsfrau.de/s/232
```

Each store can gather a bit of feedback, like stars.

### Slideshow

A special presentation mode where images and text are displayed.

----

* Edith: p/02, a/02, a/25.

----

Todo:

* /s/4 - story detail
* story formatting
* archive
* security, throttling