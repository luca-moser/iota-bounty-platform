import {action, observable} from 'mobx';

export class UIStore {
    @observable drawerOpen: boolean = false;
    @observable repoFormOpen: boolean = false;
    @observable bountyFormOpen: boolean = false;

    @observable deleteBountyModalOpen: boolean = false;
    @observable deleteRepoModalOpen: boolean = false;

    constructor() {

    }

    @action
    setDrawerOpen = (open: boolean) => this.drawerOpen = open;

    @action
    toggleDrawerOpen = () => this.drawerOpen = !this.drawerOpen;

    @action
    toggleRepoFormOpen = () => this.repoFormOpen = !this.repoFormOpen;

    @action
    setRepoFormOpen = (open: boolean) => this.repoFormOpen = open;

    @action
    toggleBountyFormOpen = () => this.bountyFormOpen = !this.bountyFormOpen;

    @action
    setBountyFormOpen = (open: boolean) => this.bountyFormOpen = open;

    @action
    setDeleteBountyModalOpen = (open: boolean) => this.deleteBountyModalOpen = open;

    @action
    setDeleteRepoModalOpen = (open: boolean) => this.deleteRepoModalOpen = open;
}

export var UIStoreInstance = new UIStore();