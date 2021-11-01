"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const broccoli_bridge_1 = __importDefault(require("broccoli-bridge"));
const broccoli_funnel_1 = __importDefault(require("broccoli-funnel"));
const broccoli_merge_trees_1 = __importDefault(require("broccoli-merge-trees"));
const broccoli_plugin_1 = __importDefault(require("broccoli-plugin"));
const broccoli_file_creator_1 = __importDefault(require("broccoli-file-creator"));
const broccoli_source_1 = require("broccoli-source");
const core_1 = __importDefault(require("@docfy/core"));
const docfy_output_template_1 = __importDefault(require("./docfy-output-template"));
const get_config_1 = __importDefault(require("./get-config"));
const utils_1 = require("./plugins/utils");
const debug_1 = __importDefault(require("debug"));
const debug = debug_1.default('@docfy/ember');
const templateOnlyComponent = `
import Component from '@glimmer/component';
export default class extends Component {}
`;
function ensureDirectoryExistence(filePath) {
    const dirname = path_1.default.dirname(filePath);
    if (fs_1.default.existsSync(dirname)) {
        return;
    }
    ensureDirectoryExistence(dirname);
    fs_1.default.mkdirSync(dirname);
}
function hasBackingJS(chunks) {
    for (let i = 0; i < chunks.length; i++) {
        const chunk = chunks[i];
        if (chunk.ext === 'js' || chunk.ext === 'ts') {
            return true;
        }
    }
    return false;
}
class DocfyBroccoli extends broccoli_plugin_1.default {
    constructor(inputNodes, options = {}) {
        super(inputNodes, options);
        this.config = options;
    }
    build() {
        return __awaiter(this, void 0, void 0, function* () {
            debug('Output Path: ', this.outputPath);
            debug('Config: ', this.config);
            const docfy = new core_1.default(this.config);
            const result = yield docfy.run(this.config.sources);
            result.content.forEach((page) => {
                const parts = [this.outputPath, 'templates', page.meta.url];
                if (page.meta.url[page.meta.url.length - 1] === '/') {
                    parts.push('index');
                }
                const fileName = `${path_1.default.join(...parts)}.hbs`;
                ensureDirectoryExistence(fileName);
                fs_1.default.writeFileSync(fileName, page.rendered);
                const demoComponents = page.pluginData.demoComponents;
                if (utils_1.isDemoComponents(demoComponents)) {
                    demoComponents.forEach((component) => {
                        component.chunks.forEach((chunk) => {
                            const chunkPath = path_1.default.join(this.outputPath, 'components', `${component.name.dashCase}.${chunk.ext}`);
                            ensureDirectoryExistence(chunkPath);
                            fs_1.default.writeFileSync(chunkPath, chunk.code);
                        });
                        if (!hasBackingJS(component.chunks)) {
                            const chunkPath = path_1.default.join(this.outputPath, 'components', `${component.name.dashCase}.js`);
                            ensureDirectoryExistence(chunkPath);
                            fs_1.default.writeFileSync(chunkPath, templateOnlyComponent);
                        }
                    });
                }
            });
            fs_1.default.writeFileSync(path_1.default.join(this.outputPath, 'docfy-output.js'), `export default ${JSON.stringify({ nested: result.nestedPageMetadata })};`);
            const urlsJsonFile = path_1.default.join(this.outputPath, 'public', 'docfy-urls.json');
            const sbJsonFile = path_1.default.join(this.outputPath, `public`, `storybook`, `stories.json`);
            ensureDirectoryExistence(urlsJsonFile);
            ensureDirectoryExistence(sbJsonFile);
            const slugify = (str) => {
              return str.replace(/[^\w\-]+/g, '-')
            };
            const kinds = {};
            const stories = {};


            const walk = function(children) {
              children.reduce((prev, item) => {
                const kind = item.name;
                prev[kind] = {
                  fileName: 1667,
                  framework: 'ember'
                };
                item.pages.reduce((prev, item) => {
                  const id = `${slugify(kind)}--${slugify(item.title)}`;
                  prev[id] = {
                    id: id,
                    name: item.title,
                    kind: kind,
                    story: item.title,
                    parameters: {
                      __id: id,
                      __isArgsStory: false
                    }
                  };
                  return prev;
                }, stories);
                walk(item.children);
                return prev;
              }, kinds);
            }
            walk(result.nestedPageMetadata.children)


            const storiesOut = {
              v: 2,
              globalParameters: {},
              kindParameters: kinds,
              stories: stories
            };

            fs_1.default.writeFileSync(sbJsonFile, `${JSON.stringify(storiesOut, null, 4)}`);
            fs_1.default.writeFileSync(urlsJsonFile, JSON.stringify(result.content.map((page) => page.meta.url)));
            result.staticAssets.forEach((asset) => {
                const dest = path_1.default.join(this.outputPath, 'public', asset.toPath);
                ensureDirectoryExistence(dest);
                fs_1.default.copyFileSync(asset.fromPath, dest);
            });
        });
    }
}
module.exports = {
    name: require('../package').name,
    docfyConfig: undefined,
    included(...args) {
        this.docfyConfig = get_config_1.default(this.project.root);
        this.bridge = new broccoli_bridge_1.default();
        this._super.included.apply(this, args);
    },
    treeForApp(tree) {
        const trees = [this._super.treeForApp.call(this, tree)];
        const inputs = [new broccoli_source_1.UnwatchedDir(this.project.root)];
        this.docfyConfig.sources.forEach((item) => {
            if (item.root && item.root !== this.project.root) {
                inputs.push(item.root);
            }
        });
        const docfyTree = new DocfyBroccoli(inputs, this.docfyConfig);
        trees.push(docfyTree);
        this.bridge.fulfill('docfy-tree', docfyTree);
        return new broccoli_merge_trees_1.default(trees, { overwrite: true });
    },
    treeForAddon(tree) {
        const trees = [this._super.treeForAddon.call(this, tree)];
        const EmberApp = require('ember-cli/lib/broccoli/ember-app'); // eslint-disable-line
        const modulePrefix = this.project.config(EmberApp.env()).modulePrefix;
        trees.push(new broccoli_file_creator_1.default('output.js', docfy_output_template_1.default(modulePrefix)));
        return new broccoli_merge_trees_1.default(trees);
    },
    treeForPublic() {
        return new broccoli_funnel_1.default(this.bridge.placeholderFor('docfy-tree'), {
            srcDir: 'public',
            destDir: './'
        });
    },
    urlsForPrember(distDir) {
        try {
            return require(path_1.default.join(distDir, 'docfy-urls.json')); // eslint-disable-line
        }
        catch (_a) {
            // empty
        }
        return [];
    }
};
