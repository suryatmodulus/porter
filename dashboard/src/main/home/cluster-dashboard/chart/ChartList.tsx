import React, { Component } from 'react';
import styled from 'styled-components';

import { Context } from '../../../../shared/Context';
import api from '../../../../shared/api';
import { ChartType, StorageType, Cluster } from '../../../../shared/types';

import Chart from './Chart';
import Loading from '../../../../components/Loading';

type PropsType = {
  currentCluster: Cluster,
  namespace: string,
  setCurrentChart: (c: ChartType) => void
};

type StateType = {
  charts: ChartType[],
  loading: boolean,
  error: boolean,
  ws: any,
};

export default class ChartList extends Component<PropsType, StateType> {
  state = {
    charts: [] as ChartType[],
    loading: false,
    error: false,
    ws : null as any
  }

  updateCharts = () => {
    let { currentCluster, currentProject, setCurrentError } = this.context;

    this.setState({ loading: true });
    setTimeout(() => {
      if (this.state.loading) {
        this.setState({ loading: false, error: true });
      }
    }, 3000);

    api.getCharts('<token>', {
      namespace: this.props.namespace,
      cluster_id: currentCluster.id,
      service_account_id: currentCluster.service_account_id,
      storage: StorageType.Secret,
      limit: 20,
      skip: 0,
      byDate: false,
      statusFilter: ['deployed', 'uninstalled', 'pending', 'pending_upgrade',
        'pending_rollback','superseded','failed']
    }, { id: currentProject.id }, (err: any, res: any) => {
        if (err) {
        console.log(err)
        setCurrentError(JSON.stringify(err));
        this.setState({ loading: false, error: true });
      } else {
        if (res.data) {
          this.setState({ charts: res.data });
        } else {
          this.setState({ charts: [] });
        }
        this.setState({ loading: false, error: false });
      }
    });
  }

  setupWebsocket = () => {
    let { currentCluster, currentProject } = this.context;
    let ws = new WebSocket(`ws://localhost:8080/api/projects/${currentProject.id}/k8s/deployment/status?cluster_id=${currentCluster.id}&service_account_id=${currentCluster.service_account_id}`)

    this.setState({ ws }, () => {
      if (!this.state.ws) return;
  
      this.state.ws.onopen = () => {
        console.log('connected to websocket')
      }
  
      this.state.ws.onmessage = (evt: MessageEvent) => {
        console.log(evt.data)
      }
  
      this.state.ws.onerror = (err: ErrorEvent) => {
        console.log(err)
      }
    })
  }

  componentDidMount() {
    this.updateCharts();
    this.setupWebsocket();
  }

  componentWillUnmount() {
    if (this.state.ws) {
      console.log('closing websocket')
      this.state.ws.close()
    }
  }

  componentDidUpdate(prevProps: PropsType) {

    // Ret2: Prevents reload when opening ClusterConfigModal
    if (prevProps.currentCluster !== this.props.currentCluster || 
      prevProps.namespace !== this.props.namespace) {
      this.updateCharts();
    }
  }

  renderChartList = () => {
    let { loading, error, charts } = this.state;

    if (loading) {
      return <LoadingWrapper><Loading /></LoadingWrapper>
    } else if (error) {
      return (
        <Placeholder>
          <i className="material-icons">error</i> Error connecting to cluster.
        </Placeholder>
      );
    } else if (charts.length === 0) {
      return (
        <Placeholder>
          <i className="material-icons">category</i> No charts found in this namespace.
        </Placeholder>
      );
    }

    return this.state.charts.map((x: ChartType, i: number) => {
      return (
        <Chart
          key={i}
          chart={x}
          setCurrentChart={this.props.setCurrentChart}
        />
      )
    })
  }


  render() {
    return (
      <StyledChartList>
        {this.renderChartList()}
      </StyledChartList>
    );
  }
}

ChartList.contextType = Context;

const Placeholder = styled.div`
  padding-top: 100px;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  color: #ffffff44;
  font-size: 14px;

  > i {
    font-size: 18px;
    margin-right: 12px;
  }
`;

const LoadingWrapper = styled.div`
  padding-top: 100px;
`;

const StyledChartList = styled.div`
  padding-bottom: 100px;
`;