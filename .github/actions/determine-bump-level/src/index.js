const core = require('@actions/core');
const { context } = require("@actions/github");
const { Octokit } = require('@octokit/rest')

const semver = require('semver');
const shell = require('shelljs');

const token = core.getInput('token', {required: true});

const baseUrl = "https://github.com";

let bumpLevel;

async function run() {
  try {
    token || core.setFailed("Missing token!");

    //const excludePrefix = core.getInput('exclusion-prefix') || '';
    var includePrefix = core.getInput('inclusion-prefix') || '';

    const owner = context.repo.owner;
    const repo = context.repo.repo;

    core.info(`Fetching tags from ${baseUrl}/${owner}/${repo}`);
    shell.exec(`git fetch ${baseUrl}/${owner}/${repo} --tags`, { silent: true });
 
    const octokit = new Octokit({ auth: token });

    const iterator = octokit.paginate.iterator("GET /repos/{owner}/{repo}/tags", {
      owner,
      repo,
      per_page: 100,
    });

    let allRepoTags = []

    for await (const {data} of iterator) {
      for (const tags of data) {
        allRepoTags = [...allRepoTags, {name: tags.name}];
      }
    }

    const matchingTags = await getMatchingTags(includePrefix);

    const latestTag = matchingTags[0];
    var previousTag = allRepoTags[0]["name"];

    if (latestTag === previousTag) {
      if (allRepoTags.length <= 1) {
        previousTag = "v0.0.0";
      } else {
        previousTag = allRepoTags[1]["name"];
      }
    }

    core.info(`All tags => ${matchingTags}`);
    core.info(`Latest tag sorted by creatordate => ${latestTag}`);
    core.info(`Previous tag sorted by semver => ${previousTag}`);

    if (!semver.valid(latestTag) || !semver.valid(previousTag)) {
      core.warning('Repository contains semver incompatible tags !');
      core.warning('You should use either inclusion-prefix or exclusion-prefix input parameter to filter them in/out!');
    }

    if (semver.compare(latestTag, previousTag) > 0) {
      bumpLevel = semver.diff(latestTag, previousTag);
      core.info(`Latest tag (${latestTag}) is higher than previous tag (${previousTag})`);
    }

    if (semver.compare(latestTag, previousTag) < 0) {
      core.warning(`Latest tag (${latestTag}) is lower than previous tag (${previousTag}), but we can handle it.`)

      const latestTagMajorDigit = includePrefix = latestTag.split('.')[0];
      const latestTagMinorDigit = latestTag.split('.')[1];

      const majorPrefixedTags = await getMatchingTags(includePrefix);
      const latestMajorPrefixedTag = majorPrefixedTags[0];
      const previousMajorPrefixedTag = majorPrefixedTags[1];

      const previousTagMajorDigit = previousMajorPrefixedTag.split('.')[0];
      const previousTagMinorDigit = previousMajorPrefixedTag.split('.')[1];

      core.info(`All tags matching "${includePrefix}" prefix => ${majorPrefixedTags}`);
      core.info(`Latest tag sorted by creatordate matching "${includePrefix}" prefix => ${latestMajorPrefixedTag}`);
      core.info(`Previous tag sorted by creatordate matching "${includePrefix}" prefix => ${previousMajorPrefixedTag}`);

      const latestTagMajorMinorDigits = includePrefix = latestTagMajorDigit.concat('.', latestTagMinorDigit);
      const previousTagMajorMinor = previousTagMajorDigit.concat('.', previousTagMinorDigit);
            
      const majorMinorPrefixedTags = await getMatchingTags(includePrefix);
      const latestMajorMajorTag = majorMinorPrefixedTags[0];
      const previousMajorMajorTag = majorMinorPrefixedTags[1];

      core.info(`All tags sorted by creatordate matching "${includePrefix}" prefix => ${majorMinorPrefixedTags}`);
      core.info(`Latest tag sorted by creatordate matching "${includePrefix}" prefix => ${latestMajorMajorTag}`);

      if (previousMajorMajorTag !== undefined) {
        core.info(`Previous tag matching "${includePrefix}" prefix => ${previousMajorMajorTag}`);
      } else {
        core.info(`There is no previous tag matching "${includePrefix}" prefix!`);
      }

      if (latestMajorPrefixedTag > previousMajorPrefixedTag) {
        if (latestTagMajorMinorDigits === previousTagMajorMinor && previousMajorMajorTag !== undefined) {
          bumpLevel = "patch";
        } else if (latestTagMajorMinorDigits > previousTagMajorMinor) {
          bumpLevel = "minor";
        } else {
          core.warning(`${latestTagMajorMinorDigits} must be greater or equal to ${previousTagMajorMinor} !`);
        }
      } else if (latestMajorPrefixedTag < previousMajorPrefixedTag) {
        bumpLevel = semver.diff(latestMajorMajorTag, previousMajorMajorTag);
      }
    }

    core.info(`Detected bump level => ${bumpLevel}`);
    core.info(`Latest tag => ${latestTag}`);
    core.info(`Previous tag => ${previousTag}`);

    core.setOutput("bump-level", bumpLevel);
    core.setOutput("latest-tag", latestTag);
    core.setOutput("previous-tag", previousTag);

  } catch (error) {
    core.setFailed(error.message);
  }
}

async function getMatchingTags(includePrefix) {

  let i = 0;
  const matchingTags = [];

  if (includePrefix !== '') {
    stdout = shell.exec(`git tag --sort=-creatordate | grep -e "${includePrefix}.*\.[0-9]"`, {silent: true});
  } else {
    stdout = shell.exec(`git tag --sort=-creatordate`, {silent: true});
  }

  const allMatchingTags = stdout.trim().split(/\r?\n/);
  const tagsLength = allMatchingTags.length;

  while (i < tagsLength) {
    var tag = allMatchingTags[i];
    if ((includePrefix !== undefined && tag.startsWith(includePrefix))) {
      matchingTags.push(tag);
    }
    i++;
  }

  return matchingTags;
}

if (require.main === module) {
  run();
}
