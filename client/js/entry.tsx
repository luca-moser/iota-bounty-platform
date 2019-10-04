declare var module;
declare var require;
require("react-hot-loader/patch");

import './../css/main.scss';

import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {BrowserRouter} from 'react-router-dom';
import {configure} from 'mobx';
import {Provider} from 'mobx-react';
import {AppContainer} from 'react-hot-loader'
import {default as App} from './comps/App';

import {AppStoreInstance as appStore} from "./stores/AppStore";
import {UIStoreInstance as uiStore} from "./stores/UIStore";
import {RepositoryStoreInstance as repoStore} from "./stores/RepositoryStore";
import {BountyStoreInstance as bountyStore} from "./stores/BountyStore";

configure({enforceActions: "observed"});

let stores = {appStore, uiStore, repoStore, bountyStore};

const render = Component => {
    ReactDOM.render(
        <AppContainer>
            <BrowserRouter>
                <Provider {...stores}>
                    <Component/>
                </Provider>
            </BrowserRouter>
        </AppContainer>,
        document.getElementById('app')
    )
}

render(App);

if (module.hot) {
    module.hot.accept()
}