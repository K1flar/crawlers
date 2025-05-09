import React, { useCallback, useEffect } from 'react';
import Dagre from '@dagrejs/dagre';
import {
    ReactFlow,
    ReactFlowProvider,
    Controls,
    Background,
    useNodesState,
    useEdgesState,
    useReactFlow,
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';

// Стандартные размеры узлов
const NODE_WIDTH = 200;
const NODE_HEIGHT = 50;

const getLayoutedElements = (nodes, edges, options) => {
    const g = new Dagre.graphlib.Graph();
    g.setGraph({
        rankdir: options.direction,
        nodesep: 50,
        ranksep: 100
    });
    g.setDefaultEdgeLabel(() => ({}));

    nodes.forEach(node => {
        g.setNode(node.id, {
            width: NODE_WIDTH,
            height: NODE_HEIGHT
        });
    });

    edges.forEach(edge => g.setEdge(edge.source, edge.target));

    Dagre.layout(g);

    return {
        nodes: nodes.map(node => {
            const { x, y } = g.node(node.id);
            return {
                ...node,
                position: {
                    x: x - NODE_WIDTH / 2,
                    y: y - NODE_HEIGHT / 2
                },
                style: { width: NODE_WIDTH }
            };
        }),
        edges
    };
};

const Graph = ({ sources }) => {
    const { nodes: initNodes, edges: initEdges } = prepareFlowData(sources)

    useEffect(() => {
        const { nodes: newNodes, edges: newEdges } = prepareFlowData(sources)
        setNodes(newNodes)
        setEdges(newEdges)
    }, [sources])

    const { fitView } = useReactFlow();
    const [nodes, setNodes, onNodesChange] = useNodesState(initNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initEdges);

    React.useEffect(() => {
        setTimeout(() => fitView({ padding: 0.5 }), 0);
    }, [nodes, fitView]);

    return (
        <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onNodeClick={(event, node) => {
                // Проверяем наличие URL в данных узла
                if (node.data?.url && node.data.url !== '#') {
                    window.open(node.data.url, '_blank', 'noopener,noreferrer');
                }
            }}
            fitView
        >
            <Controls />
            <Background />
        </ReactFlow>
    );
};

export default function ({ sources }) {
    return (
        <ReactFlowProvider>
            <Graph sources={sources} />
        </ReactFlowProvider>
    );
}

// Настройка ReactFlow
const prepareFlowData = (sources) => {
    if (!sources || !sources.length) return { nodes: [], edges: [] };

    const nodes = sources.map(source => ({
        id: `${source.id}`,
        data: { label: source.title, url: source.url },
    }));
    // Добавляем корневой узел
    nodes.unshift({
        id: 'search_engine',
        data: { label: 'Поисковая система' },
    });

    const edges = sources
        .map(source => ({
            id: `e${source.id}-${source.parentId}`,
            source: `${source.parentId || 'search_engine'}`,
            target: `${source.id}`,
        }));

    return getLayoutedElements(nodes, edges, { direction: "TB" });
};