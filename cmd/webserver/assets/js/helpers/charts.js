const defaultChartOptions = {
    chart: {
        type: 'spline',
        backgroundColor: 'rgba(0,0,0,0)',
    },
    title: {
        text: '',
    },
    subtitle: {
        text: '',
    },
    credits: {
        enabled: false,
    },
    xAxis: {
        title: {text: ''},
        type: 'datetime',
    },
    legend: {
        enabled: true,
        itemStyle: {
            color: '#28a745',
        },
        itemHiddenStyle: {
            color: '#666666',
        },
    },
    plotOptions: {
        series: {
            marker: {
                enabled: false,
            }
        }
    },
    colors: ['#28a745', '#007bff', '#e83e8c', '#ffc107', '#343a40'],
};
