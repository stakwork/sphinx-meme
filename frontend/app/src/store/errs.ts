import { observable, action } from 'mobx'
import api from '../api'

interface Err {
  message: string;
  details: string;
  code: number;
  time: string;
}

export class ErrStore {
  @observable
  errs: Err[] = []

  constructor() {
    this.fetchErrs()
  }

  @action
  async fetchErrs() {
    try {
      const r = await api.get('errs/all')
      this.errs = r
    } catch(e) {
      console.log(e)
    }
  }
}
