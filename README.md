# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.onrender.com/. Try searching for "Hamlet" to display
a set of results.

In it's current state, however, the app is in rough shape. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Your Mission

Improve the app! Think about the problem from the **user's perspective**
and prioritize your changes according to what you think is most useful.

You can approach this with a back-end, front-end, or full-stack focus.

## Evaluation

We will be primarily evaluating based on how well the search works for users. A search result with a lot of features (i.e. multi-words and mis-spellings handled), but with results that are hard to read would not be a strong submission.

## Submission

1. Fork this repository and send us a link to your fork after pushing your changes.
2. Render (render.com) hosting, the application deploys cleanly from a public url.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.

## Prioritized Changes

1. Add Boostrap styling
2. Mark string in results
3. Case-insensitive search
4. Infix/fuzzy search
5. ETL list of works - 'Contents\n' (44 works total)
    a. Hard code since this is expected to not change
    b. or find works in list of contents - 'THE SONNETS\n'... 
6. TODO: Link play
    a. External URL
    b. Better to have internal URL
7. TODO: Handle titled work contents
    a. Use titled work start and end index to search contents
    b. Better to improve seach all works with Bleve or ElasticSearch
8. TODO: Filter by work results
9. TODO: Show work
10. TODO: Multiple words search
11. TODO: Auto complete
12. TODO: Relevance score
13. TODO: ETL and search chapters
    a. Tragedy contains act and scene 
    b. Novel number/roman chapters 
    c. Poem new line and strings
14. TODO: ETL Speaker - name, dialogue, and new line
15. TODO: Pagination of results