My weekend project aimed to help me learn Go while having some fun.
Takes advantage of goroutines, channels, function signatures, etc.

# hinko: slack bot for random team picking and ASCII fun

A Slack bot written in go with its own database that can help you pick random teams (e.g. for foosball, code reviews, etc.).

It features a number of ASCII art and animation experiments, along with the ability to create custom named groups and put/get values.

The bot uses reactions to confirm commands (or say they don't work) to not spam the channel.

Try the following commands:
```help
put key value
get key
group groupname list
group groupname create @user1 @user2 @user3 ...
group groupname add @user1 @user2 @user3 ...
group groupname remove @user1 @user2 @user3 ...
randompairs @user1 @user2 @user3 ...
randompairs group
randomteams teamsize @user1 @user2 @user3 ...
randomteams teamsize group
ascii https://imageurl
shark
animate
```
![screenshot](https://github.com/tadej/hinko/blob/master/images/hinko-screen-1.png "screenshot")

![screenshot](https://github.com/tadej/hinko/blob/master/images/hinko-screen-2.png "screenshot")
