const core = require('@actions/core')
const github = require('@actions/github')
const simpleGit = require('simple-git')
const semver = require('semver')

async function run () {
  try {
    const prefix = core.getInput('prefix')
    const git = simpleGit()
    versions = (await git.listRemote(['--tags']))
      .trimEnd()
      .split('\n')
      .filter(line => {
        return line.trim().length > 0
      })
      .map(line => {
        return line.match(/^.*refs\/tags\/(.+)$/)[1]
      })
      .filter(tag => {
        return tag.startsWith(prefix)
      })
      .map(tag => {
        return tag.slice(prefix.length)
      })
      .filter(version => {
        return (version = ~/^\d/ && semver.valid(version))
      })
      .sort(semver.compare)
      .reverse()
    if (versions.length) {
      core.setOutput('tag', prefix + versions[0])
      core.setOutput('version', versions[0])
    }
  } catch (error) {
    core.setFailed(error.message)
  }
}

run()
