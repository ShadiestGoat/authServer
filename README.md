# Auth Server

## What is this?

This is a utility/example http hosting server, that has authentication for some pages.

There is 1 very specific use case for this:

1. You have html files, with no need for SSR
2. You want to do basic authentication, using the default browser login alert
3. You have a lot of pages that require different usernames/passwords
4. There is no "individual user" - users have to log in using this specific password

The only actual use case for this that I can think of (and why I even made this in the first place) is for browser based puzzle games like notpron.

## How to use

1. Place your site contents into a folder called `site`
2. Create a file called `.passwords`, and enter passwords (see syntax below)
   - NOTE: Don't create this in `site/.passwords`
3. Run `go install github.com/shadiestgoat/authServer`
4. Run `authServer` in the directory with the `.passwords` and `site` folder
   - You can set the `PORT` env variable to change the port this server runs on, by default its 3000
   - Note - there isn't any sort implementation for htaccess (unless browsers implement it) or anything like that, its just a static file host w/ authentication for some files
   - If `/foo` is requested, it will attempt both `site/foo/index.html` and `site/foo.html`

## Password File Syntax

The `.passwords` file should have the following syntax

```
/path/to/file : username here : password here
/path/to/other file : username 2 here : password 2 here
```

Notes:

- The password is optional
- There is flexibility to the paths. Use `*` as a wildcard to match 1 path section. Use `**` to match unlimited path sections. See match table below

| .       | /foo  | /bar  | /foo/bar |
| :------ | :---: | :---: | :------: |
| /foo    |   ✅   |  :x:  |   :x:    |
| /*      |   ✅   |   ✅   |   :x:    |
| /**     |   ✅   |   ✅   |    ✅     |
| /**/bar |  :x:  |  :x:  |    ✅     |

If there is a more specific or shorter pattern that is matched, then it's authentication will used.

If there is are 2 matches with later wildcard at different levels, then the one with the later one wins

```
/foo/bar/*
/foo/*/bar
```

In this example, `/foo/bar/*` wins.

## Limitations

Because I created this as a bit of a speedrun, and do not plan on doing active support for the project, there is no proper escaping. This means stuff in the `.passwords` cannot contain the string ` : `, otherwise it'll be considered the next 'section'
Also, consecutive wildcards (`*`, `**`) aren't allowed.

