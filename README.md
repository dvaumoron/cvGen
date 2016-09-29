# cvGen
generate a html cv from a yaml file using a go template.

run on [heroku](https://www.heroku.com/) [here](http://cvgenerator.herokuapp.com/)

you can use it by creating a cv.yaml ([example](https://github.com/dvaumoron/cv/blob/master/cv.yaml)) file in a cv repository in your github account then go at http://cvgenerator.herokuapp.com/githubAccountName

if you don't want to use the default template you can add a cv.html file in your repository [using go template syntax](https://golang.org/pkg/text/template/), referencing css or js file from your repository is possible using /static/githubAccountName/repositoryName/fileName

you can use a different repository with http://cvgenerator.herokuapp.com/githubAccountName/repositoryName or an other file with http://cvgenerator.herokuapp.com/githubAccountName/repositoryName/fileName

have fun
